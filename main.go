package main

import "github.com/visionmedia/docopt"
import "github.com/segmentio/go-log"
import "github.com/stathat/go"
import "strconv"
import "strings"
import "bufio"
import "time"
import "os"

var Version = "0.0.1"

const Usage = `
  Usage:
    stathat --key s
    stathat -h | --help
    stathat --version

  Options:
    -k, --key s      EZ key
    -h, --help       output help information
    -v, --version    output version

`

func main() {
	args, err := docopt.Parse(Usage, nil, true, Version, false)
	log.Check(err)

	scan(args["--key"].(string))

	stathat.WaitUntilFinished(10 * time.Second)
}

// scan stdin.
func scan(key string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if name, value, err := parse(scanner.Text()); err == nil {
			send(name, value, key)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("failed to read stdin: %s", err)
		os.Exit(1)
	}
}

// parse metric line.
func parse(line string) (name string, value float64, err error) {
	parts := strings.Split(line, "|")
	name = parts[0]
	value, err = strconv.ParseFloat(parts[1], 64)
	return
}

// send to stathat.
func send(name string, value float64, key string) {
	if strings.HasPrefix(name, "timers.") {
		if strings.HasSuffix(name, ".count") || strings.HasSuffix(name, ".sum") {
			name = strings.Replace(name, "timers.", "", 1)
			log.Info("sending timer %q %f", name, value)
			stathat.PostEZValue(name, key, value)
		}
		return
	}

	if strings.HasPrefix(name, "gauges.") {
		name = strings.Replace(name, "gauges.", "", 1)
		log.Info("sending gauge %q %f", name, value)
		stathat.PostEZValue(name, key, value)
		return
	}

	if strings.HasPrefix(name, "counts.") {
		name = strings.Replace(name, "counts.", "", 1)
		log.Info("sending count %q %d", name, int(value))
		stathat.PostEZCount(name, key, int(value))
		return
	}
}
