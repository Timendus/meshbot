package helpers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func Pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	if word == "it" {
		return "them"
	}
	return word + "s"
}

func TimeAgo(timestamp time.Time) string {
	seconds := int(time.Since(timestamp).Seconds())

	if seconds == 1 {
		return "one second"
	}
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}

	minutes := int(math.Floor(float64(seconds) / 60))
	if minutes == 1 {
		return "one minute"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}

	hours := int(math.Floor(float64(minutes) / 60))
	if hours == 1 {
		return "one hour"
	}
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	}

	days := int(math.Floor(float64(hours) / 24))
	if days == 1 {
		return "one day"
	}
	return fmt.Sprintf("%d days", days)
}

func BreakMessage(input string) []string {
	const MAX_MESSAGE_LENGTH = 200
	const MAX_LENGTH_WITH_PAGINATION = 200 - len(" [1/2]")
	input = strings.TrimSpace(input)
	messages := make([]string, 0)
	for _, message := range strings.Split(input, "<BREAK-MESSAGE>") {
		message = strings.TrimSpace(message)

		// Don't try to cut up messages that fit
		if len(message) <= MAX_MESSAGE_LENGTH {
			messages = append(messages, message)
			continue
		}

		// Cut message in parts and add pagination info to each part
		messageParts := BreakMessageAt(message, MAX_LENGTH_WITH_PAGINATION)
		for i := range messageParts {
			if len(messageParts) > 9 {
				messageParts[i] += " [" + strconv.Itoa(i+1) + "]"
			} else {
				messageParts[i] += " [" + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(messageParts)) + "]"
			}
		}

		messages = append(messages, messageParts...)
	}
	Assert(len(messages) < 1000, "What the hell are you doing creating so many messages..?")
	return messages
}

func BreakMessageAt(input string, maxlength int) []string {
	input = strings.TrimSpace(input)
	messages := make([]string, 0)
	startPtr := 0
	endPtr := 0
	resumePtr := 0

	for startPtr < len(input) {
		// Find the next (rough) place where we need to cut the input to get it
		// to fit in a message
		charEnd := startPtr + maxlength

		if charEnd >= len(input) {
			// We can fit the whole rest of the input in the message, in other
			// words: we're done
			messages = append(messages, input[startPtr:])
			break
		}

		// Find the "real" charEnd, that considers UTF-8 encoding
		// boundaries. This should walk back at most 4 bytes, and since
		// we're always considering 200 bytes at once, we should be fine.
		for !utf8.ValidString(input[startPtr:charEnd]) {
			charEnd--
		}

		// Break on the furthest newline that fits in the next message, if
		// the line after that can fit in a single message. Otherwise, break
		// on the furthest space. If neither is found, break on character.
		wordEnd := strings.LastIndex(input[startPtr:charEnd+1], " ")
		lineEnd := strings.LastIndex(input[startPtr:charEnd+1], "\n")

		nextLineEnd := strings.Index(input[charEnd:], "\n")
		if nextLineEnd == -1 {
			nextLineEnd = len(input)
		}
		nextLineLength := (nextLineEnd + charEnd) - (lineEnd + startPtr + 1)

		if lineEnd != -1 && nextLineLength <= maxlength {
			endPtr = lineEnd + startPtr
			resumePtr = endPtr + 1 // Skip the newline character
		} else if wordEnd != -1 {
			endPtr = wordEnd + startPtr
			resumePtr = endPtr + 1 // Skip the space character
		} else {
			endPtr = charEnd
			resumePtr = endPtr
		}

		messages = append(messages, input[startPtr:endPtr])
		startPtr = resumePtr
	}

	return messages
}

func Indent(s, prefix string) string {
	lines := strings.SplitAfter(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "")
}
