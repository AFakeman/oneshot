package cronparse

import (
    "bufio"
    "os"
    "strings"

    "github.com/pkg/errors"
)

type CrontabLine struct {
    Timespec string
    Command string
}

func ReadCrontab(filename string) ([]CrontabLine, error) {
    result := []CrontabLine{}

    file, err := os.Open(filename)
    if err != nil {
        return result, errors.Wrap(err, "Could not open crontab for parsing")
    }
    defer file.Close()


    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        trimmed_line := strings.Trim(scanner.Text(), "\n ")
        parsed_line, err := parseCrontabLine(trimmed_line)
        if err != nil {
            return result, err
        }
        result = append(result, parsed_line)
    }

    return result, nil
}

func parseCrontabLine(line string) (CrontabLine, error) {
    result := CrontabLine{}
    idx := nthIndexAny(line, " ", 5)
    if idx == -1 {
        return result, errors.New("Could not parse a crontab line: " + line)
    }
    result.Timespec = line[:idx]
    result.Command = line[idx+1:]
    return result, nil
}

func nthIndexAny(str string, chars string, n int) int {
    startIdx := 0;
    for i := 0; i < n; i++ {
        idx := strings.IndexAny(str[startIdx:], chars)
        if idx == -1 {
            return -1
        }
        startIdx += idx + 1
    }
    return startIdx - 1
}
