package transfer

import (
	"fmt"
	"time"
)

func Debug(msg string) {
	fmt.Printf("%s - DEBUG: %s\n", time.Now(), msg)
}

func Info(msg string) {

}
