import os
import sys
import time
import logging

from junit_xml import TestCase, TestSuite

SILENT = False

class IntegrationTest(object):
    def __init__(self, name, path, package, mem=256, timeout=10, can_crash=False, can_shutdown=False, check_state=None):
        self.name = name
        self.path = path
        self.package = package
        self.mem = mem
        self.timeout = 10
        self.can_crash = can_crash
        self.can_shutdown = can_shutdown
        self.check_state = check_state

    @classmethod
    def load(cls, path):
        from importlib import import_module
        return cls.from_module(import_module(path, ''))

    @classmethod
    def from_module(cls, m):
        return cls(
            m.name,
            m.__path__[0],
            m.package,
            getattr(m, 'mem', 256),
            getattr(m, 'timeout', 10),
            getattr(m, 'can_crash', False),
            getattr(m, 'can_shutdown', False),
            getattr(m, 'check_state', None),
        )

    @classmethod
    def discover(cls, path='.'):
        from importlib import import_module
        from operator import is_not
        from functools import partial

        dirs = [f for f in os.listdir(path) if os.path.isdir(f)]
        def to_module(f):
            try:
                return import_module(f, '')
            except ImportError:
                return None
        modules = filter(partial(is_not, None), map(to_module, dirs))
        modules = filter(lambda m: getattr(m, 'integration_test', False) == True, modules)
        tests = map(cls.from_module, modules)
        return list(tests)

    def run(self, unigornel_root):
        from tempfile import mkstemp

        prefix = 'unigornel-{0}-'.format(self.name)
        _, kernel_path = mkstemp(prefix=prefix)
        kernel_name = os.path.basename(kernel_path)

        cases = []
        def f(cases):
            # Build application
            c = self.build(unigornel_root, out=kernel_path)
            cases.append(c)
            if c.is_failure():
                return None

            # Execute application
            guest, state, c = self.execute(kernel_path, kernel_name)
            cases.append(c)
            if guest:
                log('Executed guest {0}'.format(repr(guest)))
            if state:
                log(state.console)
                log('Guest exited with state {0}'.format(hex(state.final_state)))

            if c.is_failure():
                return guest

            # Check state
            log('Checking final kernel state')
            c = self.check(state)
            cases.append(c)
            if c.is_failure():
                return guest

            return guest

        guest = f(cases)
        self.clean(guest, kernel_path)

        return TestSuite(self.name, cases)

    def build(self, unigornel_root, out=None):
        def f():
            log("Building '{0}': package {1}".format(self.name, self.package))

            gopath = os.path.join(self.path, 'go')
            app_path = os.path.join(self.path, 'go', 'src', self.package)
            app = UnigornelApp(app_path, gopath, unigornel_root)
            output = app.build(out=out)
            log(output)
            return output, None
        return build_test_case('build', self.name, f)

    def execute(self, kernel_path, kernel_name):
        k = Kernel(kernel_path, self.mem, kernel_name)
        g = None
        state = None
        exc = None
        try:
            try:
                g = k.create_paused_guest()
            except Exception as e:
                raise Exception('Could not create guest with kernel {0}'.format(kernel_path)) from e

            try:
                state = g.unpause_and_collect(timeout=self.timeout)
            except Exception as e:
                raise Exception('Could not run guest with kernel {0}'.format(kernel_path)) from e
        except Exception as e:
            exc = e

        def f():
            if exc:
                raise exc
            return state.console, None

        return g, state, build_test_case('execute', self.name, f)

    def check(self, state):
        def f():
            if not self.can_crash and bool(state.final_state & XenGuest.STATE_CRASHED):
                return None, 'The kernel crashed unexpectedly (state={0})'.format(hex(state.final_state))
            if not self.can_shutdown and bool(state.final_state & XenGuest.STATE_SHUTDOWN):
                return None, 'The kernel shutdown unexpectedly (state={0})'.format(hex(state.final_state))

            try:
                if self.check_state:
                    log('Checking state with custom function')
                    cs = self.check_state
                else:
                    log('No custom check state function defined')
                    cs = lambda _: None

                cs(state)

                log('State checks passed')
                return 'OK', None
            except Exception as e:
                raise Exception('An exception occurred while checking state') from e
        return build_test_case('check_state', self.name, f)

    def clean(self, guest, kernel_path):
        def clean_guest(g):
            if g:
                g.destroy()

        ops = [
            lambda: clean_guest(guest),
            lambda: os.remove(kernel_path),
        ]
        for o in ops:
            try:
                o()
            except:
                pass

class UnigornelApp(object):
    BUILD_CMD = './build.bash'

    def __init__(self, path, gopath, unigornel_root):
        self.path = path
        self.gopath = gopath
        self.unigornel_root = unigornel_root

    def build(self, out=None, build_all=True, verbose=True):
        from subprocess import Popen, PIPE, STDOUT

        argv = [self.BUILD_CMD, 'app', '--app', self.path]
        if out is not None: argv += ['-o', out]
        if build_all:       argv += ['-a']
        if verbose:         argv += ['-x']

        env = os.environ.copy()
        env['GOPATH'] = self.gopath

        with pushd(self.unigornel_root):
            log('[+]', ' '.join(argv))
            with Popen(argv, env=env, stdout=PIPE, stderr=STDOUT) as proc:
                b = proc.stdout.read()
                try:
                    stdout = b.decode('utf-8')
                except UnicodeDecodeError:
                    stdout = str(b)
                return stdout

                proc.wait()
                if proc.returncode != 0:
                    raise Exception('error: build.bash returned code {0}'.format(proc.returncode))

class Kernel(object):
    def __init__(self, kernel, memory, name, on_crash='preserve'):
        self.kernel = kernel
        self.memory = memory
        self.name = name
        self.on_crash = "preserve"

    def create_paused_guest(self):
        from tempfile import NamedTemporaryFile
        from subprocess import call
        with NamedTemporaryFile('w') as config:
            print('kernel = "{0}"'.format(self.kernel),     file=config)
            print('memory = {0}'.format(self.memory),       file=config)
            print('name = "{0}"'.format(self.name),         file=config)
            print('on_crash = "{0}"'.format(self.on_crash), file=config)
            config.flush()
            path = config.name

            log('Creating {0} with configuration {1}'.format(repr(self), path))
            call(['xl', 'create', '-p', '-f', path])

        try:
            guest = next(filter(lambda g: g.name == self.name, XenGuest.list()))
        except StopIteration:
            raise Exception('Failed to launch guest: could not find guest with matching name')

        assert(guest.state & guest.STATE_PAUSED)

        return guest

    def __repr__(self):
        f = [self.kernel, self.memory, self.name, self.on_crash]
        return 'Kernel(kernel={0}, memory={1}, name={2}, on_crash={3}'.format(*f)

class XenGuestState(object):
    def __init__(self, final_state, did_timeout, console):
        self.final_state = final_state
        self.did_timeout = did_timeout
        self.console = console

class XenGuest(object):
    STATE_UNKNOWN       = 0x00
    STATE_RUNNING       = 0x01
    STATE_BLOCKED       = 0x02
    STATE_PAUSED        = 0x04
    STATE_SHUTDOWN      = 0x08
    STATE_CRASHED       = 0x10
    STATE_DYING         = 0x20

    def __init__(self, name, id, mem, vcpus, raw_state, time):
        self.name = name
        self.id = id
        self.mem = mem
        self.vcpus = vcpus
        self.raw_state = raw_state
        self.time = time

    @property
    def state(self):
        return XenGuest.parse_state(self.raw_state)

    def is_shutdown(self): return bool(self.state & self.STATE_SHUTDOWN)
    def is_crashed(self):  return bool(self.state & self.STATE_CRASHED)

    def current(self):
        return next(filter(lambda g: g.id == self.id, XenGuest.list()), None)

    def destroy(self):
        from subprocess import call
        call(['xl', 'destroy', str(self.id)])

    def unpause_and_collect(self, timeout):
        from subprocess import Popen, PIPE, STDOUT, call
        with Popen(['xl', 'console', str(self.id)], stdout=PIPE, stderr=STDOUT, stdin=PIPE) as console:
            call(['xl', 'unpause', str(self.id)])

            # Wait until the unikernel is shutdown or the timeout is reached
            deadline = time.monotonic() + timeout
            did_timeout = False
            while True:
                current = self.current()
                if not current:
                    final_state = self.STATE_UNKNOWN
                    break
                elif current.is_shutdown() or current.is_crashed():
                    final_state = current.state
                    break
                elif time.monotonic() > deadline:
                    final_state = current.state
                    did_timeout = True
                    break
                time.sleep(1.0)

            time.sleep(1.0) # give console time to catch up
            console.kill()
            console.wait(timeout=1)

            b = console.stdout.read()
            try:
                stdout = b.decode('utf-8')
            except UnicodeDecodeError:
                stdout = str(b)
            return XenGuestState(final_state, did_timeout, stdout)

    @classmethod
    def from_xl_list_line(cls, line):
        import re
        name, id, mem, vcpus, state, time = re.split('\s+', line)
        return cls(name, int(id), int(mem), int(vcpus), state, float(time))

    @classmethod
    def list(cls):
        from subprocess import Popen, PIPE

        with Popen(['xl', 'list'], stdout=PIPE) as proc:
            proc.stdout.readline() # discard header
            m = map(lambda b: cls.from_xl_list_line(b.decode('utf-8').strip()), proc.stdout)
            guests = list(m)

            proc.wait()
            if proc.returncode != 0:
                raise Exception('error: xl list returned code {0}'.format(proc.returncode))

            return guests

    @staticmethod
    def parse_state(raw_state):
        mask = 0
        r, b, p, s, c, d = list(raw_state)
        if r == 'r': mask |= XenGuest.STATE_RUNNING
        if b == 'b': mask |= XenGuest.STATE_BLOCKED
        if p == 'p': mask |= XenGuest.STATE_PAUSED
        if s == 's': mask |= XenGuest.STATE_SHUTDOWN
        if c == 'c': mask |= XenGuest.STATE_CRASHED
        if d == 'd': mask |= XenGuest.STATE_DYING
        return mask

    def __repr__(self):
        f = [self.name, self.id, self.mem, self.vcpus, self.raw_state, self.time]
        return 'XenGuest(name={0}, id={1}, mem={2}, vcpus={3}, raw_state={4}, time={5}'.format(*f)

def main(unigornel_root, tests, **kwargs):
    if tests is not None:
        try:
            all_tests = list(map(IntegrationTest.load, tests))
        except ImportError as e:
            raise Exception('Could not load a test specified on the command line') from e
    else:
        all_tests = IntegrationTest.discover()

    n = len(all_tests)
    test_suites = []
    log('Running {0} test(s)'.format(n))
    for i, test in enumerate(all_tests):
        log("----------")
        log("Running test '{0}' ({1}/{2})".format(test.name, i+1, n))
        ts = test.run(unigornel_root)
        test_suites.append(ts)
    return test_suites


def parse(argv, envs):
    from argparse import ArgumentParser
    parser = ArgumentParser(description='Execute the integration tests')
    parser.add_argument('tests', metavar='test', type=str, nargs='*', help='only run specified tests')
    parser.add_argument('-s', dest='silent', action='store_true', help='silent mode')
    parser.add_argument('--junit', type=str, help='write junit xml to specified file')

    args = parser.parse_args(argv)
    try:
        unigornel_root = envs['UNIGORNEL_ROOT']
    except KeyError:
        raise Exception('Environment variable UNIGORNEL_ROOT is not set')

    if args.silent:
        global SILENT
        SILENT = True

    return {
        'unigornel_root' : unigornel_root,
        'tests' : args.tests if len(args.tests) != 0 else None,
        'junit' : args.junit,
    }

def log(*args, **kwargs):
    if not SILENT:
        print(*args, **kwargs)

def pushd(dirname):
    from os import chdir, getcwd
    from os.path import realpath
    class PushdContext(object):
        cmd = None
        original_dir = None

        def __init__(self, dirname):
            self.cwd = realpath(dirname)

        def __enter__(self):
            self.original_dir = getcwd()
            chdir(self.cwd)
            return self

        def __exit__(self, type, value, tb):
            chdir(self.original_dir)
    return PushdContext(dirname)

def build_test_case(name, classname, f):
    from traceback import format_exc
    start = time.monotonic()
    try:
        output, failure = f()
    except Exception:
        output = format_exc()
        failure = "An unexpected exception occurred."
        log('{0}:'.format(failure))
        log(output)

    elapsed = time.monotonic() - start
    tc = TestCase(name, classname, elapsed, output)
    if failure is not None:
        tc.add_failure_info(failure)
    return tc

if __name__ == "__main__":
    args = parse(sys.argv[1:], os.environ)
    suites = main(**args)

    junit = args['junit']
    if junit:
        log('Writing', len(suites), ' test suites to', junit)
        with open(junit, 'w') as f:
            TestSuite.to_file(f, suites)
