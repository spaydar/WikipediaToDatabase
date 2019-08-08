package WikipediaToDatabase

import (
	"testing"
)

func TestMainMethod(t *testing.T) {
	err := main()
	if err != nil {
		t.Error("Error in connecting to database", err)
		return
	}
}
