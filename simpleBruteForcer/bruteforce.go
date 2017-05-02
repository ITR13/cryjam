package bruteforce

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type format struct {
	prefix                 string
	alphabet               string
	currentLetter          []int
	minLetters, maxLetters int
	next                   *format
}

type BruteForcer struct {
	format              format
	lowercaseOnly       bool
	seperator, suffix   string
	o                   io.Writer
	finished, running   bool
	tasks, threads, tpt int
	lock                *sync.Mutex
	closed              bool
}

func Parse(s string) (arguments []string) {
	escaped := false
	current := ""
	state := 0
	for _, c := range s {
		switch c {
		case '\\':
			if state == 0 {
				state = 1
			}

			if escaped {
				current += "\\"
			}
			escaped = !escaped
		case '"':
			switch state {
			case 0:
				state = 2
			case 1:
				if escaped {
					escaped = false
				}
				current += "\""
			case 2:
				if escaped {
					current += "\""
					escaped = false
				} else {
					arguments = append(arguments, current)
					current = ""
					state = 0
				}
			}
		case ' ':
			switch state {
			case 1:
				if escaped {
					escaped = false
					current += " "
				} else {
					arguments = append(arguments, current)
					current = ""
					state = 0
				}
			case 2:
				if escaped {
					escaped = false
				}
				current += " "
			}
		default:
			if escaped {
				escaped = false
			}
			if state == 0 {
				state = 1
			}
			current += string(c)
		}
	}
	if state != 0 {
		arguments = append(arguments, current)
	}
	return
}

func (f *format) Next(addOne bool) string {
	s := f.prefix
	if f.alphabet != "" && f.maxLetters != 0 {
		for i := range f.currentLetter {
			if addOne {
				f.currentLetter[i]++
				if f.currentLetter[i] == len(f.alphabet) {
					f.currentLetter[i] = 0
				} else {
					addOne = false
				}
			}
			s += string(f.alphabet[f.currentLetter[i]])
		}
		if addOne {
			if len(f.currentLetter) == f.maxLetters {
				f.currentLetter = make([]int, f.minLetters)
				if f.next == nil {
					return ""
				}
			} else {
				f.currentLetter = make([]int, len(f.currentLetter)+1)
				addOne = false
			}
		}
	}
	if f.next != nil {
		n := f.next.Next(addOne)
		if n == "" {
			return ""
		}
		return s + n
	} else if addOne {
		return ""
	}
	return s
}

func createFormat(input [][4]string) (root *format, err error,
	otherFeedback string) {
	root = &format{}
	current := root
	previous := root
	for i := range input {
		switch input[i][0] {
		case "-ss":
			fallthrough
		case "-setstring":
			current.prefix += input[i][1]
		case "-a":
			fallthrough
		case "-alphabet":
			current.alphabet = input[i][1]
			var min, max int
			min, err = strconv.Atoi(input[i][2])
			if err != nil {
				root = nil
				return
			}
			if min < 1 {
				otherFeedback += "min was less than 1, so setting to 1\n"
				min = 1
			}
			max, err = strconv.Atoi(input[i][3])
			if err != nil {
				root = nil
				return
			}
			if max < min {
				otherFeedback += "max was less than min, so setting to min\n"
				max = min
			}
			current.currentLetter = make([]int, min)
			current.minLetters, current.maxLetters = min, max
			next := &format{}
			current.next = next
			previous, current = current, next
		default:
			otherFeedback += "Unknown input: \"" + input[i][0] + "\"\n"
		}
	}
	if current.alphabet == "" && current.prefix == "" {
		previous.next = nil
	}
	return
}

func Affirm(input [][4]string, output [2]string, lowercaseOnly bool,
	seperator string, threads, tasks int) (message string) {
	if threads < 1 {
		message += "Threads less than 1, so setting to 1\n"
		threads = 1
	}
	if tasks < 1 {
		message += "Tasks less than 1, so setting to 1\n"
		tasks = 1
	}
	if seperator == "" {
		message += "Seperator empty, so setting to \\n\n"
		seperator = "\n"
	}
	f, err, mes := createFormat(input)
	if err != nil {
		return err.Error()
	}
	if mes != "" {
		message += mes
	}
	end := f
	foundAlphabet := false
	if f != nil {
		for end.next != nil {
			if end.alphabet != "" && end.minLetters != 0 {
				foundAlphabet = true
				break
			}
			end = end.next
		}
		if end.alphabet != "" && end.minLetters != 0 {
			foundAlphabet = true
		}
	} else {
		f = &format{}
		end = f
	}

	if !foundAlphabet {
		message += "Found no alphabet so setting to abc\n"
		end.alphabet = "abcdefghijklmnopqrstuvwxyz"
		end.minLetters = 1
		end.maxLetters = 6
	}

	for i := f; i.next != nil; i = i.next {
		if i.alphabet != "" {
			if strings.Contains(i.alphabet, seperator) {
				message += "Alphabet \"" + i.alphabet + "\" contains " +
					"seperator \"" + seperator + "\", so unexpected " +
					" behaviour may occur\n"
			}
		}
	}

	if output[0] == "" {
		message += "output-prefix is empty, so outputting to stdout\n"
	}
	return
}
func GetBruteForcer(input [][4]string, output [2]string, lowercaseOnly bool,
	seperator string, threads, tasks int) *BruteForcer {
	if threads < 1 {
		threads = 1
	}
	if tasks < 1 {
		tasks = 1
	}

	if seperator == "" {
		seperator = "\n"
	}

	f, _, _ := createFormat(input)
	end := f
	foundAlphabet := false
	if f != nil {
		for end.next != nil {
			if end.alphabet != "" && end.minLetters != 0 {
				foundAlphabet = true
				break
			}
			end = end.next
		}
		if end.alphabet != "" && end.minLetters != 0 {
			foundAlphabet = true
		}
	} else {
		f = &format{}
		end = f
	}

	if !foundAlphabet {
		end.alphabet = "abcdefghijklmnopqrstuvwxyz"
		end.minLetters = 1
		end.maxLetters = 6
	}

	br := BruteForcer{*f, lowercaseOnly, seperator, output[1],
		nil, false, false, 0, threads, tasks, &sync.Mutex{}, false}
	if output[0] == "" {
		br.o = os.Stdout
	} else {
		c := Parse(output[0])
		command := exec.Command(c[0], c[1:]...)
		reader, writer := io.Pipe()
		br.o = writer
		command.Stdin = reader
		command.Stdout = os.Stdout
		command.Run()
	}
	return &br
}

func (br *BruteForcer) BruteForce(threadless bool) {
	br.lock.Lock()
	if br.running {
		br.lock.Unlock()
		return
	}
	br.running = true
	br.lock.Unlock()
	for i := &br.format; i != nil; i = i.next {
		if i.alphabet != "" && len(i.currentLetter) > 0 {
			i.currentLetter[0] = -1
			break
		}
	}

	channel := make(chan []string)

	if threadless {
		for i := 0; i < br.threads; i++ {
			br.thread(channel)
		}
	} else {
		for i := 0; i < br.threads; i++ {
			go br.thread(channel)
		}
	}

	c := 0
	for !br.finished {
		for ; c < br.tasks; c++ {
			s := <-channel
			for i := 0; i < len(s); i++ {
				br.o.Write([]byte(s[i]))
			}
		}
	}
	br.o.Write([]byte(br.suffix))
	br.lock.Lock()
	br.finished = false
	br.running = false
	br.lock.Unlock()
}

func (br *BruteForcer) thread(c chan []string) {
	task := br.getTask()
	for len(task) != 0 {
		results := make([]string, len(task))
		for i := range task {
			args := Parse(task[i])
			command := exec.Command(args[0], args[1:]...)
			r, w, err := os.Pipe()
			if err != nil {
				results[i] = err.Error()
				continue
			}
			defer r.Close()
			defer w.Close()
			command.Stdout = w
			err = command.Run()
			if err != nil {
				results[i] = err.Error()
			} else {
				reader := bufio.NewReader(r)
				results[i], _ = reader.ReadString('\n')
			}
		}
		go func() { c <- results }()
		task = br.getTask()
	}
}

func (br *BruteForcer) getTask() []string {
	br.lock.Lock()
	defer br.lock.Unlock()
	if br.finished || !br.running {
		return make([]string, 0)
	}

	s := make([]string, br.tpt)
	for i := range s {
		f := br.format.Next(true)
		if f == "" {
			br.finished = true
			s = s[:i]
			break
		}
		s[i] = f
	}

	if len(s) > 0 {
		br.tasks++
	}

	return s
}

func (br *BruteForcer) Close() {
	br.lock.Lock()
	defer br.lock.Unlock()
	i := 0
	for br.running {
		if br.closed {
			return
		}
		br.lock.Unlock()
		i++
		time.Sleep(time.Second)
		br.lock.Lock()
	}
	br.running = true
	br.finished = true
	switch t := br.o.(type) {
	case *io.PipeWriter:
		t.Close()
	}
	br.closed = true
}
