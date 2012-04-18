// Copyright 2011 Robert Bloomquist

//	NAME
//	 	wc - count lines, words, and characters
//	
//	SYNOPSIS
//		wc [ -lwceb ] [ file ... ]
//	
//	DESCRIPTION
//		Wc writes to standard output a tally of lines, words, and
//		characters found in each file, assumed to be text in UTF
//		format.  If no files are named, standard input is read.  One
//		line is output per file.  If several files are specified, an
//		additional line is written giving totals.
//	
//		`Words' are maximal sequences of characters separated by
//		blanks, tabs and newlines.
//	
//		Counts are output in the same order as the listing of the
//		option letters lwceb; select lines, words, UTF characters,
//		erroneously-encoded characters, and bytes, respectively.  If
//		no options are given, lines, words, and characters are
//		counted.
//	
//	BUGS
//		The Unicode Standard has many blank characters scattered
//		through it, but wc looks for only ASCII space, tab, and new-
//		line.
//	
//		Wc should have options to count suboptimal UTF codes and
//		bytes that cannot occur in any UTF code.

package main

import (
	"flag"
	"fmt"
	"os"
)

const NBUF = 8 * 1024

// Command-line flags
var (
	lflag = flag.Bool("l", false, "print number of lines")
	wflag = flag.Bool("w", false, "print number of words")
	cflag = flag.Bool("c", false, "print number of characters")
	eflag = flag.Bool("e", false, "print number of erroneously-encoded characters")
	bflag = flag.Bool("b", false, "print number of bytes")
)

type counter struct {
	lines, words, chars, errors, bytes uint64
}

func report(c *counter, name string) {
	var s string
	if *lflag {
		s += fmt.Sprintf("%7d", c.lines)
	}
	if *wflag {
		s += fmt.Sprintf("%7d", c.words)
	}
	if *cflag {
		s += fmt.Sprintf("%7d", c.chars)
	}
	if *eflag {
		s += fmt.Sprintf("%7d", c.errors)
	}
	if *bflag {
		s += fmt.Sprintf("%7d", c.bytes)
	}
	s += " " + name + "\n"
	os.Stdout.WriteString(s)
}

// How it works.  Start in statesp.  Each time we read a character,
// increment various counts, and do state transitions according to the
// following table.  If we're not in statesp or statewd when done, the
// file ends with a partial rune.
//        |                character
//  state |09,20| 0a  |00-7f|80-bf|c0-df|e0-ef|f0-ff
// -------+-----+-----+-----+-----+-----+-----+-----
// statesp|ASP  |ASPN |AWDW |AWDWX|AC2W |AC3W |AWDWX
// statewd|ASP  |ASPN |AWD  |AWDX |AC2  |AC3  |AWDX
// statec2|ASPX |ASPNX|AWDX |AWDR |AC2X |AC3X |AWDX
// statec3|ASPX |ASPNX|AWDX |AC2R |AC2X |AC3X |AWDX

const ( // actions
	AC2   = iota // enter statec2
	AC2R         // enter statec2, don't count a rune
	AC2W         // enter statec2, count a word
	AC2X         // enter statec2, count a bad rune
	AC3          // enter statec3
	AC3W         // enter statec3, count a word
	AC3X         // enter statec3, count a bad rune
	ASP          // enter statesp
	ASPN         // enter statesp, count a newline
	ASPNX        // enter statesp, count a newline, count a bad rune
	ASPX         // enter statesp, count a bad rune
	AWD          // enter statewd
	AWDR         // enter statewd, don't count a rune
	AWDW         // enter statewd, count a word
	AWDWX        // enter statewd, count a word, count a bad rune
	AWDX         // enter statewd, count a bad rune
)

var statesp = [256]byte{ // looking for the start of a word
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 00-07
	AWDW, ASP, ASPN, AWDW, AWDW, AWDW, AWDW, AWDW, // 08-0f
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 10-17
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 18-1f
	ASP, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 20-27
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 28-2f
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 30-37
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 38-3f
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 40-47
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 48-4f
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 50-57
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 58-5f
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 60-67
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 68-6f
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 70-77
	AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, AWDW, // 78-7f
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // 80-87
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // 88-8f
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // 90-97
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // 98-9f
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // a0-a7
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // a8-af
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // b0-b7
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // b8-bf
	AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, // c0-c7
	AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, // c8-cf
	AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, // d0-d7
	AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, AC2W, // d8-df
	AC3W, AC3W, AC3W, AC3W, AC3W, AC3W, AC3W, AC3W, // e0-e7
	AC3W, AC3W, AC3W, AC3W, AC3W, AC3W, AC3W, AC3W, // e8-ef
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // f0-f7
	AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, AWDWX, // f8-ff
}

var statewd = [256]byte{ // looking for the next character in a word
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 00-07
	AWD, ASP, ASPN, AWD, AWD, AWD, AWD, AWD, // 08-0f
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 10-17
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 18-1f
	ASP, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 20-27
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 28-2f
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 30-37
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 38-3f
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 40-47
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 48-4f
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 50-57
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 58-5f
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 60-67
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 68-6f
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 70-77
	AWD, AWD, AWD, AWD, AWD, AWD, AWD, AWD, // 78-7f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 80-87
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 88-8f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 90-97
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 98-9f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // a0-a7
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // a8-af
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // b0-b7
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // b8-bf
	AC2, AC2, AC2, AC2, AC2, AC2, AC2, AC2, // c0-c7
	AC2, AC2, AC2, AC2, AC2, AC2, AC2, AC2, // c8-cf
	AC2, AC2, AC2, AC2, AC2, AC2, AC2, AC2, // d0-d7
	AC2, AC2, AC2, AC2, AC2, AC2, AC2, AC2, // d8-df
	AC3, AC3, AC3, AC3, AC3, AC3, AC3, AC3, // e0-e7
	AC3, AC3, AC3, AC3, AC3, AC3, AC3, AC3, // e8-ef
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // f0-f7
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // f8-ff
}

var statec2 = [256]byte{ // looking for 10xxxxxx to complete a rune
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 00-07
	AWDX, ASPX, ASPNX, AWDX, AWDX, AWDX, AWDX, AWDX, // 08-0f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 10-17
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 18-1f
	ASPX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 20-27
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 28-2f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 30-37
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 38-3f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 40-47
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 48-4f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 50-57
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 58-5f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 60-67
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 68-6f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 70-77
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 78-7f
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // 80-87
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // 88-8f
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // 90-97
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // 98-9f
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // a0-a7
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // a8-af
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // b0-b7
	AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, AWDR, // b8-bf
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // c0-c7
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // c8-cf
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // d0-d7
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // d8-df
	AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, // e0-e7
	AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, // e8-ef
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // f0-f7
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // f8-ff
}

var statec3 = [256]byte{ // looking for 10xxxxxx,10xxxxxx to complete a rune
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 00-07
	AWDX, ASPX, ASPNX, AWDX, AWDX, AWDX, AWDX, AWDX, // 08-0f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 10-17
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 18-1f
	ASPX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 20-27
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 28-2f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 30-37
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 38-3f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 40-47
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 48-4f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 50-57
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 58-5f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 60-67
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 68-6f
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 70-77
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // 78-7f
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // 80-87
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // 88-8f
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // 90-97
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // 98-9f
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // a0-a7
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // a8-af
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // b0-b7 
	AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, AC2R, // b8-bf
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // c0-c7
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // c8-cf
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // d0-d7
	AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, AC2X, // d8-df
	AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, // e0-e7
	AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, AC3X, // e8-ef
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // f0-f7
	AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, AWDX, // f8-ff
}

func count(n *counter, f *os.File) {
	state := statesp[:]
	var buf [NBUF]byte
	n.lines = 0
	n.words = 0
	n.chars = 0
	n.errors = 0
	n.bytes = 0
	for {
		nr, er := f.Read(buf[:])
		if nr == 0 {
			break
		}
		if nr < 0 {
			fmt.Fprintf(os.Stderr, "wc: error reading from %s: %s\n", f.Name(), er.Error())
			os.Exit(1)
		}
		n.bytes += uint64(nr)
		n.chars += uint64(nr) // might be too large, gets decreased later
		for i := 0; i < nr; i++ {
			switch state[buf[i]] {
			case AC2:
				state = statec2[:]
			case AC2R:
				state = statec2[:]
				n.chars--
			case AC2W:
				state = statec2[:]
				n.words++
			case AC2X:
				state = statec2[:]
				n.errors++
			case AC3:
				state = statec3[:]
			case AC3W:
				state = statec3[:]
				n.words++
			case AC3X:
				state = statec3[:]
				n.errors++
			case ASP:
				state = statesp[:]
			case ASPN:
				state = statesp[:]
				n.lines++
			case ASPNX:
				state = statesp[:]
				n.lines++
				n.errors++
			case ASPX:
				state = statesp[:]
				n.errors++
			case AWD:
				state = statewd[:]
			case AWDR:
				state = statewd[:]
				n.chars--
			case AWDW:
				state = statewd[:]
				n.words++
			case AWDWX:
				state = statewd[:]
				n.words++
				n.errors++
			case AWDX:
				state = statewd[:]
				n.errors++
			}
		}
	}
}

func main() {
	n := new(counter)
	t := new(counter)

	flag.Parse()
	if flag.NFlag() == 0 {
		*lflag = true
		*wflag = true
		*cflag = true
	}

	if flag.NArg() == 0 {
		count(n, os.Stdin)
		report(n, "")
	}
	for i := 0; i < flag.NArg(); i++ {
		f, err := os.Open(flag.Arg(i))
		defer f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wc: can't open %s: error %s\n", flag.Arg(i), err)
			os.Exit(1)
		}
		count(n, f)
		t.lines += n.lines
		t.words += n.words
		t.chars += n.chars
		t.errors += n.errors
		t.bytes += n.bytes
		report(n, flag.Arg(i))
	}
	if flag.NArg() > 1 {
		report(t, "total")
	}
}