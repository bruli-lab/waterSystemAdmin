package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
	pongo3 "github.com/flosch/pongo2/v6"
)

func RemoveTemperatureProgram(set *pongo3.TemplateSet, removeSvc *programs.RemoveTemperature, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.ErrorContext(r.Context(), "Method not allowed", slog.String("method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tpl, err := set.FromFile("redirect_message.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		context := map[string]any{
			"page": "programs",
		}

		err = removingTemperatureProgram(r, removeSvc)
		switch {
		case err != nil:
			log.ErrorContext(r.Context(), "error removing temperature program",
				slog.String("error", err.Error()))
			context["error"] = err.Error()
		default:
			context["success"] = true
		}

		if err = tpl.ExecuteWriter(context, w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func removingTemperatureProgram(r *http.Request, svc *programs.RemoveTemperature) error {
	temp, err := strconv.Atoi(r.PathValue("temperature"))
	if err != nil {
		return fmt.Errorf("invalid temperature: %w", err)
	}
	if err = svc.Remove(r.Context(), temp); err != nil {
		return fmt.Errorf("faild remove temperature program: %w", err)
	}
	return nil
}
