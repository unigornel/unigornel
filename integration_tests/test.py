import os
import sys
import logging

SILENT = False

class IntegrationTest(object):
    def __init__(self, name, path, package):
        self.name = name
        self.path = path
        self.package = package

    @classmethod
    def load(cls, path):
        from importlib import import_module
        return cls.from_module(import_module(path, ''))

    @classmethod
    def from_module(cls, m):
        return cls(m.name, m.__path__[0], m.package)

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
        self.build(unigornel_root)

    def build(self, unigornel_root):
        log("Building '{0}': package {1}".format(self.name, self.package))

        gopath = os.path.join(self.path, 'go')
        app_path = os.path.join(self.path, 'go', 'src', self.package)
        app = UnigornelApp(app_path, gopath, unigornel_root)
        app.build()

    def execute(self):
        pass

    def clean(self):
        pass

class UnigornelApp(object):
    BUILD_CMD = './build.bash'

    def __init__(self, path, gopath, unigornel_root):
        self.path = path
        self.gopath = gopath
        self.unigornel_root = unigornel_root

    def build(self, out=None, build_all=True, verbose=True):
        from subprocess import Popen, PIPE

        argv = [self.BUILD_CMD, 'app', '--app', self.path]
        if out is not None: argv += ['-o', out]
        if build_all:       argv += ['-a']
        if verbose:         argv += ['-x']

        env = os.environ.copy()
        env['GOPATH'] = self.gopath

        with pushd(self.unigornel_root):
            with Popen(argv, env=env):
                pass

def main(unigornel_root, tests):
    if tests is not None:
        try:
            all_tests = list(map(IntegrationTest.load, tests))
        except ImportError as e:
            raise Exception('Could not load a test specified on the command line') from e
    else:
        all_tests = IntegrationTest.discover()

    n = len(all_tests)
    log('Running {0} test(s)'.format(n))
    for i, test in enumerate(all_tests):
        log("----------")
        log("Running test '{0}' ({1}/{2})".format(test.name, i+1, n))
        test.run(unigornel_root)

def parse(argv, envs):
    from argparse import ArgumentParser
    parser = ArgumentParser(description='Execute the integration tests')
    parser.add_argument('tests', metavar='test', type=str, nargs='*', help='only run specified tests')
    parser.add_argument('-s', dest='silent', action='store_true', help='silent mode')

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

if __name__ == "__main__":
    args = parse(sys.argv[1:], os.environ)
    main(**args)
