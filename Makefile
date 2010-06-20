include $(GOROOT)/src/Make.$(GOARCH)

TARG=main
GOFILES=main.go regexp.go simple.go submatch.go util.go data.go

include $(GOROOT)/src/Make.cmd
