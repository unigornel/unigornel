integration_test = True
name = 'HelloWorldTest'
package = 'helloworld'

can_crash = True
can_shutdown = True

def check_state(state):
    assert('Hello World!' in state.console)
    assert('not in console' in state.console)
