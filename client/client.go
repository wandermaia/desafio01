package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Struct que será utilizada para armazenar o valor da cotação recebida do server
type Bid struct {
	Bid string `json:"bid"`
}

func main() {

	// Definição do timeout do contexto em 300 ms
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Consulta do valor da cotação fornecida pelo server
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro na consulta ao servidor!\n")
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var ColetaBid Bid
	err = json.Unmarshal(body, &ColetaBid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer a parse da resposta:\n")
		panic(err)
	}

	// Executando a função para gravar a cotação no arquivo
	err = GravaArquivo(ColetaBid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gravar os dados no arquivo: \n")
		panic(err)
	}

	fmt.Printf("Cotação: %v. O Valor foi salvo no arquivo 'cotacao.txt'. \n", ColetaBid.Bid)

}

// Função para gravar os dados no arquivo cotacao.txt
func GravaArquivo(bid Bid) (err error) {

	// Criando o arquivo
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s\n", bid.Bid))
	if err != nil {
		return err
	}

	return nil

}
