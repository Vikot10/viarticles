package get

import (
	"net/http"

	"github.com/Vikot10/viarticles/internal/dto"
)

type articleGetter interface {
	GetArticleByID(id string) (*dto.Article, error)
}

func Handler(rw http.ResponseWriter, req *http.Request, getter articleGetter) {

}
