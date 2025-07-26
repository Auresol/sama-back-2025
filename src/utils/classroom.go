package utils

import (
	"strconv"
	"strings"
)

func ClassroomSplit(classroom string) (uint, uint) {
	classroomSplit := strings.Split(classroom, "/")
	class, _ := strconv.Atoi(classroomSplit[0])
	room, _ := strconv.Atoi(classroomSplit[1])

	return uint(class), uint(room)
}
