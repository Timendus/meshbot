package meshtastic

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	if word == "it" {
		return "them"
	}
	return word + "s"
}
