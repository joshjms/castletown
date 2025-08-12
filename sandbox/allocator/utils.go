package allocator

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"
)

func getSubuid() (uint32, uint32, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, 0, err
	}

	f, err := os.Open("/etc/subuid")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		var subuidUser string
		var subuidStart, subuidSize uint32

		splitLine := strings.Split(line, ":")
		if len(splitLine) != 3 {
			continue
		}

		subuidUser = splitLine[0]
		subuidStart = parseUint32(splitLine[1])
		subuidSize = parseUint32(splitLine[2])

		if subuidUser == currentUser.Username {
			return subuidStart, subuidSize, nil
		}
	}

	return 0, 0, fmt.Errorf("no subuid found for user %s", currentUser.Username)
}

func getSubgid() (uint32, uint32, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, 0, err
	}

	f, err := os.Open("/etc/subgid")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		var subgidUser string
		var subgidStart, subgidSize uint32

		splitLine := strings.Split(line, ":")
		if len(splitLine) != 3 {
			continue
		}

		subgidUser = splitLine[0]
		subgidStart = parseUint32(splitLine[1])
		subgidSize = parseUint32(splitLine[2])

		if subgidUser == currentUser.Username {
			return subgidStart, subgidSize, nil
		}
	}
	return 0, 0, fmt.Errorf("no subgid found for user %s", currentUser.Username)
}

func parseUint32(s string) uint32 {
	var i uint32
	fmt.Sscanf(s, "%d", &i)
	return i
}
