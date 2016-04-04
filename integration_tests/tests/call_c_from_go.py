integration_test = True

can_crash = True
can_shutdown = True

def check_state(state):
    assert 'Hello from test' in state.console
    assert 'Hello World!' in state.console
