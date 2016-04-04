.PHONY: all clean

ifeq ($(MINIOS_ROOT),)
MINIOS_ROOT=../../..
endif

HEADERS=go_pthread.h runtime.h syscalls.h types.h waittypes.h list.h experimental.h
HEADERS_SOURCE=$(patsubst %,$(MINIOS_ROOT)/include/%,$(HEADERS))
HEADERS_LOCAL=$(patsubst %,include/%,$(HEADERS))

all:

include/%.h: $(MINIOS_ROOT)/include/%.h include
	cp $< $@

include:
	mkdir -p include
	
include/mini-os: $(HEADERS_LOCAL)
	rm -f include/mini-os
	ln -s ./ include/mini-os
