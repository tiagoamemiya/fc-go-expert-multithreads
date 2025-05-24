package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type AddressInfo struct {
	Data   string
	Source SourceFinder
}

type SourceFinder struct {
	Name string
	Url  string
}

func AddressFinder(zipcode int, source SourceFinder) AddressInfo {
	finderUrl := strings.Replace(source.Url, "{zipcode}", strconv.Itoa(zipcode), -1)
	req, err := http.Get(finderUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer a request: %v\n", err)
		panic(err)
	}

	defer req.Body.Close()
	res, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler a resposta: %v\n", err)
		panic(err)
	}

	var data map[string]interface{}

	err = json.Unmarshal(res, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer o parse da resposta: %v\n", err)
		panic(err)
	}

	var info AddressInfo
	resString, _ := json.Marshal(data)
	info.Data = string(resString)
	info.Source = source
	return info
}

func main() {
	brazilApi := make(chan AddressInfo)
	viaCep := make(chan AddressInfo)

	go func() {
		source := SourceFinder{Name: "BrasilApi", Url: "https://brasilapi.com.br/api/cep/v1/{zipcode}"}
		res := AddressFinder(11050030, source)
		brazilApi <- res
	}()

	go func() {
		source := SourceFinder{Name: "iaCep", Url: "https://viacep.com.br/ws/{zipcode}/json/"}
		res := AddressFinder(11050030, source)
		viaCep <- res
	}()

	select {
	case info := <-brazilApi:
		fmt.Printf("Source %s\n", info.Source.Name)
		fmt.Print(info.Data)
	case info := <-viaCep:
		fmt.Printf("Source %s\n", info.Source.Name)
		fmt.Print(info.Data)
	case <-time.After(time.Second * 1):
		fmt.Println("Timeout")
	}

}
