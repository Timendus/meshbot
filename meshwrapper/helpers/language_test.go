package helpers

import (
	"testing"
)

const TWO_HUNDRED_CHARS = "Helloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahelloahe"
const TWO_HUNDRED_CHAR_WORDS = "Hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello hello he"

func TestPluralize(t *testing.T) {
	Assert(Pluralize("it", 0) == "them", "Pluralize(it, 0) == them")
	Assert(Pluralize("it", 1) == "it", "Pluralize(it, 1) == it")
	Assert(Pluralize("it", 2) == "them", "Pluralize(it, 2) == them")
	Assert(Pluralize("it", 3) == "them", "Pluralize(it, 3) == them")
	Assert(Pluralize("it", 4) == "them", "Pluralize(it, 4) == them")

	Assert(Pluralize("message", 0) == "messages", "Pluralize(message, 0) == messages")
	Assert(Pluralize("message", 1) == "message", "Pluralize(message, 1) == message")
	Assert(Pluralize("message", 2) == "messages", "Pluralize(message, 2) == messages")
	Assert(Pluralize("message", 3) == "messages", "Pluralize(message, 3) == messages")
	Assert(Pluralize("message", 4) == "messages", "Pluralize(message, 4) == messages")
}

func TestMessagePagination(t *testing.T) {
	AssertBreaking(t,
		"",
		[]string{""},
	)

	AssertBreaking(t,
		"Hello",
		[]string{"Hello"},
	)

	AssertBreaking(t,
		"Hello\nHello",
		[]string{"Hello\nHello"},
	)

	AssertBreaking(t,
		"Hello<BREAK-MESSAGE>Hello",
		[]string{"Hello", "Hello"},
	)

	AssertBreaking(t,
		TWO_HUNDRED_CHARS,
		[]string{TWO_HUNDRED_CHARS},
	)

	AssertBreaking(t,
		TWO_HUNDRED_CHAR_WORDS,
		[]string{TWO_HUNDRED_CHAR_WORDS},
	)

	AssertBreaking(t,
		TWO_HUNDRED_CHAR_WORDS+"<BREAK-MESSAGE>   Hello  <BREAK-MESSAGE>  Hello",
		[]string{
			TWO_HUNDRED_CHAR_WORDS,
			"Hello",
			"Hello",
		},
	)

	AssertBreaking(t,
		TWO_HUNDRED_CHARS+"a",
		[]string{
			TWO_HUNDRED_CHARS[:len(TWO_HUNDRED_CHARS)-6] + " [1/2]",
			"lloahea [2/2]",
		},
	)

	AssertBreaking(t,
		`🤖👋 Hey there! I understand these commands:

✉️ Message box - An answering machine for Meshtastic
- INBOX: Check your inbox
- NEW: Get new messages
- OLD: Get old messages
- CLEAR: Clear old messages
- SEND: Leave a message (SEND <id> <message>)

📶 Signal reporting - Know what I'm seeing
- /SIGNAL: Get signal report (/SIGNAL [<id>])`,
		[]string{
			`🤖👋 Hey there! I understand these commands:

✉️ Message box - An answering machine for Meshtastic
- INBOX: Check your inbox
- NEW: Get new messages
- OLD: Get old messages [1/2]`,
			`- CLEAR: Clear old messages
- SEND: Leave a message (SEND <id> <message>)

📶 Signal reporting - Know what I'm seeing
- /SIGNAL: Get signal report (/SIGNAL [<id>]) [2/2]`,
		},
	)

	AssertBreaking(t, `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed volutpat dolor rhoncus, fringilla mauris sed, tristique elit. Nulla facilisi. Phasellus orci tortor, finibus sed eleifend eget, malesuada at nisl. Nullam viverra libero sit amet metus fermentum, et vestibulum nulla fermentum. Aenean rutrum sed urna in efficitur. Curabitur ac nulla ut ante accumsan facilisis a quis arcu. Ut dapibus dolor lectus, eget semper lectus venenatis finibus. Etiam dapibus pulvinar ex, a dictum sem gravida quis. Praesent suscipit sem orci, a ultricies ligula luctus a. Fusce porta sem sed nibh feugiat condimentum.

Donec id tortor in ligula scelerisque imperdiet nec eu libero. Morbi congue hendrerit arcu, id rhoncus elit placerat ac. Vivamus efficitur quis nisi a aliquam. Quisque auctor aliquam interdum. Suspendisse facilisis lacus non efficitur ultrices. Cras a lacus dui. Maecenas neque risus, molestie ultrices velit eget, iaculis egestas erat. Praesent pharetra congue justo, id condimentum lectus pharetra sit amet. Proin a sagittis dolor, a interdum sem. Aenean at erat id augue hendrerit efficitur sed ac odio. Phasellus quis malesuada nulla. Sed id consequat erat. Curabitur sagittis eros nec sem facilisis, sed cursus purus pretium. Proin fermentum ut purus nec ultricies.
		`,
		[]string{
			`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed volutpat dolor rhoncus, fringilla mauris sed, tristique elit. Nulla facilisi. Phasellus orci tortor, finibus sed eleifend eget, [1/7]`,
			`malesuada at nisl. Nullam viverra libero sit amet metus fermentum, et vestibulum nulla fermentum. Aenean rutrum sed urna in efficitur. Curabitur ac nulla ut ante accumsan facilisis a quis arcu. [2/7]`,
			`Ut dapibus dolor lectus, eget semper lectus venenatis finibus. Etiam dapibus pulvinar ex, a dictum sem gravida quis. Praesent suscipit sem orci, a ultricies ligula luctus a. Fusce porta sem sed [3/7]`,
			`nibh feugiat condimentum.

Donec id tortor in ligula scelerisque imperdiet nec eu libero. Morbi congue hendrerit arcu, id rhoncus elit placerat ac. Vivamus efficitur quis nisi a aliquam. Quisque [4/7]`,
			`auctor aliquam interdum. Suspendisse facilisis lacus non efficitur ultrices. Cras a lacus dui. Maecenas neque risus, molestie ultrices velit eget, iaculis egestas erat. Praesent pharetra congue [5/7]`,
			`justo, id condimentum lectus pharetra sit amet. Proin a sagittis dolor, a interdum sem. Aenean at erat id augue hendrerit efficitur sed ac odio. Phasellus quis malesuada nulla. Sed id consequat [6/7]`,
			`erat. Curabitur sagittis eros nec sem facilisis, sed cursus purus pretium. Proin fermentum ut purus nec ultricies. [7/7]`,
		},
	)

	// If we have more than 9 message parts, we omit the total number (so the
	// pagination numbers always stays withtin three characters)
	AssertBreaking(t, `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed volutpat dolor rhoncus, fringilla mauris sed, tristique elit. Nulla facilisi. Phasellus orci tortor, finibus sed eleifend eget, malesuada at nisl. Nullam viverra libero sit amet metus fermentum, et vestibulum nulla fermentum. Aenean rutrum sed urna in efficitur. Curabitur ac nulla ut ante accumsan facilisis a quis arcu. Ut dapibus dolor lectus, eget semper lectus venenatis finibus. Etiam dapibus pulvinar ex, a dictum sem gravida quis. Praesent suscipit sem orci, a ultricies ligula luctus a. Fusce porta sem sed nibh feugiat condimentum.

Donec id tortor in ligula scelerisque imperdiet nec eu libero. Morbi congue hendrerit arcu, id rhoncus elit placerat ac. Vivamus efficitur quis nisi a aliquam. Quisque auctor aliquam interdum. Suspendisse facilisis lacus non efficitur ultrices. Cras a lacus dui. Maecenas neque risus, molestie ultrices velit eget, iaculis egestas erat. Praesent pharetra congue justo, id condimentum lectus pharetra sit amet. Proin a sagittis dolor, a interdum sem. Aenean at erat id augue hendrerit efficitur sed ac odio. Phasellus quis malesuada nulla. Sed id consequat erat. Curabitur sagittis eros nec sem facilisis, sed cursus purus pretium. Proin fermentum ut purus nec ultricies.

Etiam aliquet neque mollis, commodo sapien non, tincidunt risus. Maecenas sed quam iaculis, vehicula nisi eu, elementum risus. Ut lacinia scelerisque dolor id pharetra. Ut rutrum, mi id viverra commodo, urna leo malesuada sapien, ac aliquam quam orci ac nulla. Nunc feugiat diam id erat luctus dictum. Duis arcu leo, rhoncus id ipsum vitae, auctor mollis est. Donec laoreet rutrum eros a imperdiet. Duis convallis purus eu auctor venenatis. In commodo orci vitae ullamcorper suscipit. Vestibulum eleifend, augue in laoreet eleifend, nisi odio convallis mi, vitae tristique risus risus ut nulla. Duis blandit metus eu diam vehicula, eu vestibulum neque iaculis. Morbi nec viverra nunc. In mollis vitae sem nec placerat.
		`,
		[]string{
			`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed volutpat dolor rhoncus, fringilla mauris sed, tristique elit. Nulla facilisi. Phasellus orci tortor, finibus sed eleifend eget, [1]`,
			`malesuada at nisl. Nullam viverra libero sit amet metus fermentum, et vestibulum nulla fermentum. Aenean rutrum sed urna in efficitur. Curabitur ac nulla ut ante accumsan facilisis a quis arcu. [2]`,
			`Ut dapibus dolor lectus, eget semper lectus venenatis finibus. Etiam dapibus pulvinar ex, a dictum sem gravida quis. Praesent suscipit sem orci, a ultricies ligula luctus a. Fusce porta sem sed [3]`,
			`nibh feugiat condimentum.

Donec id tortor in ligula scelerisque imperdiet nec eu libero. Morbi congue hendrerit arcu, id rhoncus elit placerat ac. Vivamus efficitur quis nisi a aliquam. Quisque [4]`,
			`auctor aliquam interdum. Suspendisse facilisis lacus non efficitur ultrices. Cras a lacus dui. Maecenas neque risus, molestie ultrices velit eget, iaculis egestas erat. Praesent pharetra congue [5]`,
			`justo, id condimentum lectus pharetra sit amet. Proin a sagittis dolor, a interdum sem. Aenean at erat id augue hendrerit efficitur sed ac odio. Phasellus quis malesuada nulla. Sed id consequat [6]`,
			`erat. Curabitur sagittis eros nec sem facilisis, sed cursus purus pretium. Proin fermentum ut purus nec ultricies.

Etiam aliquet neque mollis, commodo sapien non, tincidunt risus. Maecenas sed [7]`,
			"quam iaculis, vehicula nisi eu, elementum risus. Ut lacinia scelerisque dolor id pharetra. Ut rutrum, mi id viverra commodo, urna leo malesuada sapien, ac aliquam quam orci ac nulla. Nunc [8]",
			"feugiat diam id erat luctus dictum. Duis arcu leo, rhoncus id ipsum vitae, auctor mollis est. Donec laoreet rutrum eros a imperdiet. Duis convallis purus eu auctor venenatis. In commodo orci [9]",
			"vitae ullamcorper suscipit. Vestibulum eleifend, augue in laoreet eleifend, nisi odio convallis mi, vitae tristique risus risus ut nulla. Duis blandit metus eu diam vehicula, eu vestibulum neque [10]",
			"iaculis. Morbi nec viverra nunc. In mollis vitae sem nec placerat. [11]",
		},
	)
}

func TestBreakMessage(t *testing.T) {
	AssertBreakingAt(t,
		"",
		[]string{""},
	)

	AssertBreakingAt(t,
		"Hello",
		[]string{"Hello"},
	)

	AssertBreakingAt(t,
		"Hello\nHello",
		[]string{"Hello\nHello"},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHARS,
		[]string{TWO_HUNDRED_CHARS},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHAR_WORDS,
		[]string{TWO_HUNDRED_CHAR_WORDS},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHARS+"a",
		[]string{
			TWO_HUNDRED_CHARS,
			"a",
		},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHARS+"📟!",
		[]string{
			TWO_HUNDRED_CHARS,
			"📟!",
		},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHARS[:len(TWO_HUNDRED_CHARS)-4]+"📟!",
		[]string{
			TWO_HUNDRED_CHARS[:len(TWO_HUNDRED_CHARS)-4] + "📟",
			"!",
		},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHARS[:len(TWO_HUNDRED_CHARS)-2]+"📟!",
		[]string{
			TWO_HUNDRED_CHARS[:len(TWO_HUNDRED_CHARS)-2],
			"📟!",
		},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHAR_WORDS+"y",
		[]string{
			TWO_HUNDRED_CHAR_WORDS[:len(TWO_HUNDRED_CHAR_WORDS)-3],
			"hey",
		},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHAR_WORDS+` Testing:
   - A thing here
   - And another one`,
		[]string{
			TWO_HUNDRED_CHAR_WORDS,
			`Testing:
   - A thing here
   - And another one`,
		},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHAR_WORDS+`
Testing:
   - A thing here
   - And another one`,
		[]string{
			TWO_HUNDRED_CHAR_WORDS,
			`Testing:
   - A thing here
   - And another one`,
		},
	)

	AssertBreakingAt(t,
		TWO_HUNDRED_CHAR_WORDS+"\n"+TWO_HUNDRED_CHAR_WORDS+"\n"+"Working!\n",
		[]string{
			TWO_HUNDRED_CHAR_WORDS,
			TWO_HUNDRED_CHAR_WORDS,
			"Working!",
		},
	)

	AssertBreakingAt(t,
		`🤖👋 Hey there! I understand these commands:

✉️ Message box - An answering machine for Meshtastic
- INBOX: Check your inbox
- NEW: Get new messages
- OLD: Get old messages
- CLEAR: Clear old messages
- SEND: Leave a message (SEND <id> <message>)

📶 Signal reporting - Know what I'm seeing
- /SIGNAL: Get signal report (/SIGNAL [<id>])`,
		[]string{
			`🤖👋 Hey there! I understand these commands:

✉️ Message box - An answering machine for Meshtastic
- INBOX: Check your inbox
- NEW: Get new messages
- OLD: Get old messages`,
			`- CLEAR: Clear old messages
- SEND: Leave a message (SEND <id> <message>)

📶 Signal reporting - Know what I'm seeing
- /SIGNAL: Get signal report (/SIGNAL [<id>])`,
		},
	)

	AssertBreakingAt(t,
		`lsddjfksdjfhskjfhakfjhakfashflkshv fshdis uh sdkjvh aichua ssklvjhsd ivuhsv kjsdhvd iasvha vjhvl kajvh iusv sivhkjfh aklvh siuvh svhakjhslfgslkvh sdich ivhajkfhs kjvgsliv iuhv skjvhslhvljshlksjhvisudv svhlsiuvhsvjhslkcavshvluishv hslivhslkjvhskjchsldkvhjd kshv kjshv skjhv slhvkjshvlks hvlkshvlskjvh skvhsv kjshv sdfjsl fkslfj sdlfj slfj ksldfj ljdljskf lksdflkjsf lkj sdkfj lskjf sdkfjls sdgkjlk sf klj fslkj lkjdflsdkjglsdfjk lsjf slkfj lshf slkj klsjflshflksj sfkhslfh`,
		[]string{
			`lsddjfksdjfhskjfhakfjhakfashflkshv fshdis uh sdkjvh aichua ssklvjhsd ivuhsv kjsdhvd iasvha vjhvl kajvh iusv sivhkjfh aklvh siuvh svhakjhslfgslkvh sdich ivhajkfhs kjvgsliv iuhv skjvhslhvljshlksjhvisudv`,
			`svhlsiuvhsvjhslkcavshvluishv hslivhslkjvhskjchsldkvhjd kshv kjshv skjhv slhvkjshvlks hvlkshvlskjvh skvhsv kjshv sdfjsl fkslfj sdlfj slfj ksldfj ljdljskf lksdflkjsf lkj sdkfj lskjf sdkfjls sdgkjlk sf`,
			`klj fslkj lkjdflsdkjglsdfjk lsjf slkfj lshf slkj klsjflshflksj sfkhslfh`,
		},
	)

	AssertBreakingAt(t, `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed volutpat dolor rhoncus, fringilla mauris sed, tristique elit. Nulla facilisi. Phasellus orci tortor, finibus sed eleifend eget, malesuada at nisl. Nullam viverra libero sit amet metus fermentum, et vestibulum nulla fermentum. Aenean rutrum sed urna in efficitur. Curabitur ac nulla ut ante accumsan facilisis a quis arcu. Ut dapibus dolor lectus, eget semper lectus venenatis finibus. Etiam dapibus pulvinar ex, a dictum sem gravida quis. Praesent suscipit sem orci, a ultricies ligula luctus a. Fusce porta sem sed nibh feugiat condimentum.

Donec id tortor in ligula scelerisque imperdiet nec eu libero. Morbi congue hendrerit arcu, id rhoncus elit placerat ac. Vivamus efficitur quis nisi a aliquam. Quisque auctor aliquam interdum. Suspendisse facilisis lacus non efficitur ultrices. Cras a lacus dui. Maecenas neque risus, molestie ultrices velit eget, iaculis egestas erat. Praesent pharetra congue justo, id condimentum lectus pharetra sit amet. Proin a sagittis dolor, a interdum sem. Aenean at erat id augue hendrerit efficitur sed ac odio. Phasellus quis malesuada nulla. Sed id consequat erat. Curabitur sagittis eros nec sem facilisis, sed cursus purus pretium. Proin fermentum ut purus nec ultricies.
		`,
		[]string{
			`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed volutpat dolor rhoncus, fringilla mauris sed, tristique elit. Nulla facilisi. Phasellus orci tortor, finibus sed eleifend eget, malesuada`,
			`at nisl. Nullam viverra libero sit amet metus fermentum, et vestibulum nulla fermentum. Aenean rutrum sed urna in efficitur. Curabitur ac nulla ut ante accumsan facilisis a quis arcu. Ut dapibus dolor`,
			`lectus, eget semper lectus venenatis finibus. Etiam dapibus pulvinar ex, a dictum sem gravida quis. Praesent suscipit sem orci, a ultricies ligula luctus a. Fusce porta sem sed nibh feugiat`,
			`condimentum.

Donec id tortor in ligula scelerisque imperdiet nec eu libero. Morbi congue hendrerit arcu, id rhoncus elit placerat ac. Vivamus efficitur quis nisi a aliquam. Quisque auctor aliquam`,
			`interdum. Suspendisse facilisis lacus non efficitur ultrices. Cras a lacus dui. Maecenas neque risus, molestie ultrices velit eget, iaculis egestas erat. Praesent pharetra congue justo, id condimentum`,
			`lectus pharetra sit amet. Proin a sagittis dolor, a interdum sem. Aenean at erat id augue hendrerit efficitur sed ac odio. Phasellus quis malesuada nulla. Sed id consequat erat. Curabitur sagittis`,
			`eros nec sem facilisis, sed cursus purus pretium. Proin fermentum ut purus nec ultricies.`,
		},
	)
}

func AssertBreaking(t *testing.T, message string, expected []string) {
	parts := BreakMessage(message)

	for i, part := range parts {
		if i >= len(expected) {
			t.Errorf(`Got more messages than I expected:
["%v"]`, part)
			break
		}
		if part != expected[i] {
			t.Errorf(`Expected message %d to be:
["%v"]
But got:
["%v"]`, i+1, expected[i], part)
		}
	}
}

func AssertBreakingAt(t *testing.T, message string, expected []string) {
	parts := BreakMessageAt(message, 200)

	for i, part := range parts {
		if i >= len(expected) {
			t.Errorf(`Got more messages than I expected:
["%v"]`, part)
			break
		}
		if part != expected[i] {
			t.Errorf(`Expected message %d to be:
["%v"]
But got:
["%v"]`, i+1, expected[i], part)
		}
	}
}
