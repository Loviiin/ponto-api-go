package funcoes

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFuncoes_StrParaUint(t *testing.T) {
	f := NewFuncoes()

	tests := []struct {
		name    string
		input   string
		want    uint
		wantErr bool
	}{
		{
			name:    "Número válido",
			input:   "123",
			want:    123,
			wantErr: false,
		},
		{
			name:    "Zero",
			input:   "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "Número negativo (erro)",
			input:   "-1",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Texto não numérico",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Vazio",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.StrParaUint(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFuncoes_GetUintIDFromContext(t *testing.T) {
	f := NewFuncoes()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		setupCtx  func(*gin.Context)
		key       string
		want      uint
		wantErr   bool
		errSubstr string
	}{
		{
			name: "ID válido no contexto",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "123")
			},
			key:     "user_id",
			want:    123,
			wantErr: false,
		},
		{
			name: "Chave não existe",
			setupCtx: func(c *gin.Context) {
				// nada
			},
			key:       "user_id",
			want:      0,
			wantErr:   true,
			errSubstr: "não encontrado no contexto",
		},
		{
			name: "Valor não é string",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", 123) // int, não string
			},
			key:       "user_id",
			want:      0,
			wantErr:   true,
			errSubstr: "não é uma string",
		},
		{
			name: "Valor string inválida",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "abc")
			},
			key:       "user_id",
			want:      0,
			wantErr:   true,
			errSubstr: "não é um número válido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupCtx(c)

			got, err := f.GetUintIDFromContext(c, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
