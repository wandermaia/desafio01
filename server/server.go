package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
)

// Struct para armazenar os dados da coleta da api. Também será utilizada para a criação da tabela no banco de dados pelo gorm.
type CotacaoDolar struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

// Struct auxiliar utilizada para armazenar apenas o Bid e codificar em json que será retornado para o cliente.
type BidOnly struct {
	Bid string `json:"bid"`
}

func main() {

	// Embora não tenha sido solcitada, optei por criar um home apenas para validar o funcionamento do server.
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/cotacao", BuscaCotacaoHandler)
	log.Println("Servidor iniciado!")
	http.ListenAndServe(":8080", nil)

}

// Função utilizada para validar o funcionamento do server.
// Exibe apenas uma mensagem informativa e um código 404 para paths diferentes de "/" ou "/cotacao".
// Registra uma mensagem no log informando que houve uma requisição no home
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Erro no Path!")
		return
	}

	w.Write([]byte("Funcionando!"))
	log.Println("Request executada na home!")
}

// Função handler para buscar a cotação
func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Request Iniciada")
	log.Println("Coletando os dados no site 'https://economia.awesomeapi.com.br/json/last/USD-BRL'.")
	cotacao, err := BuscaCotacao()
	if err != nil {
		log.Printf("Erro ao coletar os dados: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("Persistindo os dados no banco SQLite")
	err = GravaDados(cotacao)
	if err != nil {
		log.Printf("Erro ao gravar no banco de dados: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Segregando apenas o valor do Bid para o Response e realizando o encode
	bidOnly := BidOnly{Bid: cotacao.Bid}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidOnly)
	w.Header().Set("Content-Type", "application/json")
	log.Println("Request Finalizada")

}

// Função que realiza a busca da cotação do Dollar
func BuscaCotacao() (*CotacaoDolar, error) {

	// Definição do timeout do contexto em 200 ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Consulta do valor da cotação fornecida pelo server
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer a requisição:\n")
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// O retorno dos dados na consulta da api é json. Porém, o resultado que precisamos é um outro json, que é o um dos elementos do json principal.
	// Devido a isso, foi necessário realizar uma "conversão" para que pudessemos utilizar esse json no struct corretamente. Essa conversão vai facilitar a persistência
	// no banco de dados que será realizada mais a frente no código.
	// A variável cotacao é utilizada para retornar os dados no final da função e o map data é utilizado para realizar a "conversão"
	var cotacao CotacaoDolar
	var data map[string]CotacaoDolar

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	cotacao = data["USDBRL"]
	return &cotacao, nil

}

// Função que grava os dados na tabela no banco de dados SQLite
func GravaDados(dado *CotacaoDolar) error {
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	db.AutoMigrate(&CotacaoDolar{})

	// Definindo o contexto com timeout de 10ms
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	//Inserir a cotação no banco de dados utilizando o contexto configurado acima
	if err := db.WithContext(ctx).Create(&dado).Error; err != nil {
		return err
	}

	return nil
}
