package generator

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
)

func GenerateClientInput(client_num int, exp_num int, value_num []int, des string) {
	// Ensure the folder exists
	err := os.MkdirAll(des, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}

	for i := 0; i < client_num; i++ {
		// List to store data objects
		dataList := make([]map[string]interface{}, 0)

		// Generate data objects dynamically
		for j := 0; j < exp_num; j++ {
			expID := "exp" + strconv.Itoa(j+1)
			secrets := generateSecrets(value_num[j])

			data := map[string]interface{}{
				"Exp_ID":  expID,
				"Secrets": secrets,
			}

			dataList = append(dataList, data)
		}

		// Open a new file for writing
		fileName := fmt.Sprintf("input_c%s.json", strconv.Itoa(i+1))
		filePath := filepath.Join(des, fileName)

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		// Create a JSON encoder
		encoder := json.NewEncoder(file)

		// Write data objects as a list in the file
		err = encoder.Encode(dataList)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}

	}

}

func generateSecrets(size int) []int {
	secrets := make([]int, 0)
	for j := 0; j < size; j++ {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(2)))
		if err != nil {
			panic(err)
		}
		secrets = append(secrets, int(value.Int64()))
	}
	return secrets
}
