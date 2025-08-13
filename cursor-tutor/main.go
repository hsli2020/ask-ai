import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupFile renames the given filename to a new name with the same extension and optionally moves it to the "./archive" folder.
// The new name is suffixed with the file's creation timestamp.
func BackupFile(filename string, moveToArchive bool) error {
	// Get the file's extension
	extension := filepath.Ext(filename)

	// Get the file's creation time
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}
	creationTime := fileInfo.ModTime()

	// Create the new name with the timestamp suffix
	newName := fmt.Sprintf("%s_%s%s", filename[:len(filename)-len(extension)], creationTime.Format("20060102150405"), extension)

	// Add the archive path if moveToArchive is true
	if moveToArchive {
		newName = fmt.Sprintf("./archive/%s", newName)
	}

	// Rename and optionally move the file
	return os.Rename(filename, newName)
}

// GenerateNextSerialNumber generates the next serial number based on the current serial number
var serialNumber int

func GenerateNextSerialNumber() int {
	return serialNumber++
}


// 10-20,22,25-30
func ParseNumberRange(input string) ([]int, error) {
    var result []int
    ranges := strings.Split(input, ",")
    for _, r := range ranges {
        if strings.Contains(r, "-") {
            bounds := strings.Split(r, "-")
            if len(bounds) != 2 {
                return nil, fmt.Errorf("invalid range: %s", r)
            }
            start, err := strconv.Atoi(bounds[0])
            if err != nil {
                return nil, fmt.Errorf("invalid range: %s", r)
            }
            end, err := strconv.Atoi(bounds[1])
            if err != nil {
                return nil, fmt.Errorf("invalid range: %s", r)
            }
            for i := start; i <= end; i++ {
                result = append(result, i)
            }
        } else {
            num, err := strconv.Atoi(r)
            if err != nil {
                return nil, fmt.Errorf("invalid number: %s", r)
            }
            result = append(result, num)
        }
    }
    return result, nil
}

func ExtractOrderNumber(input string) string {
    re := regexp.MustCompile(`Your order (\d{3})-(\d{6})-(\d{3}) has been shipped`)
    match := re.FindStringSubmatch(input)
    if len(match) != 4 {
        return ""
    }
    orderNumber := fmt.Sprintf("%s-%s-%s", match[1], match[2], match[3])
    return orderNumber
}

func Dump(input interface{}) {
	json, _ := json.Marshal(input)
	fmt.Println(string(json))
}


