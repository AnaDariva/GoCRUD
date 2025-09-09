package main

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

func sanitizeDigits(s string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(s, "")
}

func main() {
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("erro ao abrir banco:", err)
	}
	if err := db.AutoMigrate(&Pessoa{}); err != nil {
		log.Fatal("erro ao migrar:", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// CREATE
	r.POST("/pessoas", func(c *gin.Context) {
		var p Pessoa
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p.CPF = sanitizeDigits(p.CPF)
		p.CEP = sanitizeDigits(p.CEP)
		p.Telefone = strings.TrimSpace(p.Telefone)

		if p.Nome == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nome é obrigatório"})
			return
		}
		if p.CPF != "" && len(p.CPF) != 11 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cpf deve ter 11 dígitos"})
			return
		}

		if err := db.Create(&p).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, p)
	})

	// LISTAR
	r.GET("/pessoas", func(c *gin.Context) {
		var ps []Pessoa
		if err := db.Order("id").Find(&ps).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ps)
	})

	// BUSCAR POR ID
	r.GET("/pessoas/:id", func(c *gin.Context) {
		var p Pessoa
		if err := db.First(&p, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "pessoa não encontrada"})
			return
		}
		c.JSON(http.StatusOK, p)
	})

	// ATUALIZAR
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

		if err := db.Save(&p).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, p)
	})

	// REMOVER
	r.DELETE("/pessoas/:id", func(c *gin.Context) {
		if err := db.Delete(&Pessoa{}, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal("erro ao subir servidor:", err)
	}
}
