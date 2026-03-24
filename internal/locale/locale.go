package locale

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed en.json
var enJSON []byte

//go:embed zh.json
var zhJSON []byte

var messages map[string]string

func Init() {
	data := enJSON
	if strings.HasPrefix(detectLang(), "zh") {
		data = zhJSON
	}
	_ = json.Unmarshal(data, &messages)
}

// T looks up key and formats it with optional args.
// Falls back to the key itself if not found.
func T(key string, args ...any) string {
	s, ok := messages[key]
	if !ok {
		return key
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

func detectLang() string {
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return "en"
}
