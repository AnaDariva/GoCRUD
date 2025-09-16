package main

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin" // Framework web (semelhante a usar bibliotecas HTTP em C, mas muito mais simples)
	"gorm.io/driver/sqlite"
	"gorm.io/gorm" // ORM (mapeia struct Go em tabelas no banco, C seria feito via SQL manual)
)

// Modelo Pessoa
// parecido com C, consjunto de campos agrupos.
// Em C seria um `struct Pessoa { ... };`, mas em Go temos tags adicionais
// que ajudam na serialização JSON, na validação (Gin) e no mapeamento do banco (GORM).
type Pessoa struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Nome        string `json:"nome" binding:"required"`
	CPF         string `json:"cpf" gorm:"uniqueIndex"`
	Telefone    string `json:"telefone"`
	Rua         string `json:"rua"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	CEP         string `json:"cep"`
	Cidade      string `json:"cidade"`
	Estado      string `json:"estado"`
}

// Função utilitária para limpar caracteres não numéricos.
// Em C seria necessário percorrer a string com um loop e `isdigit()`;
// no Go usamos expressões prontas.
func sanitizeDigits(s string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(s, "")
}

func main() {
	// Conectar ao banco SQLite usando GORM.
	// Em C: precisaria chamar `sqlite3_open` e escrever SQL na mão.
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("erro ao abrir banco:", err) //erros nao sao exceçoes
	}
	// Migrar automaticamente o schema com base no struct Pessoa.
	// Em C teríamos que executar `CREATE TABLE` manualmente.
	if err := db.AutoMigrate(&Pessoa{}); err != nil {
		log.Fatal("erro ao migrar:", err)
	}
	// Criação do roteador Gin (similar a iniciar um servidor HTTP).
	r := gin.Default()

	// Endpoint verifica se a API está no ar.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// CREATE → Cadastrar nova pessoa
	r.POST("/pessoas", func(c *gin.Context) {
		var p Pessoa
		// Faz o bind automático do JSON para a struct Pessoa
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p.CPF = sanitizeDigits(p.CPF)
		p.CEP = sanitizeDigits(p.CEP)
		p.Telefone = strings.TrimSpace(p.Telefone)

		//Regras validação
		if p.Nome == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nome é obrigatório"})
			return
		}
		if p.CPF != "" && len(p.CPF) != 11 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cpf deve ter 11 dígitos"})
			return
		}

		// Insere no banco
		// Em C seria necessário montar o SQL: INSERT INTO pessoas ()
		if err := db.Create(&p).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, p)
	})

	// Listar as pessoas
	r.GET("/pessoas", func(c *gin.Context) {
		var ps []Pessoa
		if err := db.Order("id").Find(&ps).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ps)
	})

	// Buscar pessoa por ID
	r.GET("/pessoas/:id", func(c *gin.Context) {
		var p Pessoa
		// Em C: SELECT * FROM pessoas WHERE id = ?
		if err := db.First(&p, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "pessoa não encontrada"})
			return
		}
		c.JSON(http.StatusOK, p)
	})

	// Atualizar pessoa
	r.PUT("/pessoas/:id", func(c *gin.Context) {
		var p Pessoa
		if err := db.First(&p, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "pessoa não encontrada"})
			return
		}

		var in Pessoa
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		in.CPF = sanitizeDigits(in.CPF)
		in.CEP = sanitizeDigits(in.CEP)
		p.Nome = in.Nome
		p.CPF = in.CPF
		p.Telefone = in.Telefone
		p.Rua = in.Rua
		p.Numero = in.Numero
		p.Complemento = in.Complemento
		p.Bairro = in.Bairro
		p.CEP = in.CEP
		p.Cidade = in.Cidade
		p.Estado = in.Estado

		// Em C seria preciso montar UPDATE ..SET ...
		if err := db.Save(&p).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, p)
	})

	// Remover pessoa
	r.DELETE("/pessoas/:id", func(c *gin.Context) {
		// Em C seria "DELETE FROM pessoas WHERE id = ?"
		if err := db.Delete(&Pessoa{}, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// interface web principal
	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})
	r.Static("/static", "./public")

	// Iniciar servidor
	// Em C seria necessário abrir socket, bind, listen e loop de accept.
	if err := r.Run(":8080"); err != nil {
		log.Fatal("erro ao subir servidor:", err)
	}
}
