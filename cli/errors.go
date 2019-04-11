package cli

import (
	"fmt"
)

var ErrAlreadyWorkingOnAStory = fmt.Errorf("already working on a story")
var ErrCommandRequiresAnArgument = fmt.Errorf("this command requires an argument")
var ErrCommandTakesNoArguments = fmt.Errorf("this command takes no arguments")
var ErrNotWorkingOnAStory = fmt.Errorf("not working on a story")
