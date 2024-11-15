package delete

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
	resp "shortener/internal/lib/api/response"
	"shortener/internal/lib/logger/sl"
	"shortener/internal/storage"
)

type Response struct {
	resp.Response
}

type URLRemover interface {
	DeleteURL(ctx context.Context, alias string) error
}

func New(log *slog.Logger, urlRemover URLRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		ctx := r.Context()
		isAdmin := ctx.Value("isAdmin")
		if isAdmin == nil {
			log.Info("user is not an admin")
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, resp.Error("forbidden"))

			return
		}
		if !isAdmin.(bool) {
			log.Info("user is not an admin")
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, resp.Error("forbidden"))

			return
		}

		err := urlRemover.DeleteURL(r.Context(), alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("alias doesn't exists", slog.String("url", alias))

			render.JSON(w, r, resp.Error("alias doesn't exists"))

			return
		}
		if err != nil {
			log.Error("invalid alias", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("alias deleted")

		render.JSON(w, r, Response{
			Response: resp.OK(),
		})

	}
}
