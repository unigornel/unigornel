integration_test = True

can_crash = True
can_shutdown = True

stdin = "Unigornel\n".encode('UTF-8')

def check_state(state):
    assert "Hello, what's your name? Hello, Unigornel" in state.console
