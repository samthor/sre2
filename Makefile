include $(GOROOT)/src/Make.$(GOARCH)

TARG=main
GOFILES=main.go
DEPS=sre2

include $(GOROOT)/src/Make.cmd
