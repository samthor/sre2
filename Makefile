include $(GOROOT)/src/Make.$(GOARCH)

TARG=main
GOFILES=main.go regexp.go util.go

include $(GOROOT)/src/Make.cmd
