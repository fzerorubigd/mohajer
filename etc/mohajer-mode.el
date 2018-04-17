(require 'generic-x)

(define-generic-mode
    'mohajer-mode                          ;; name of the mode
  '("#")                                  ;; comments delimiter
  '("table" "schema" "column" "index" "option" "primary" "engine")
  '(
    ("\\(`[^`]+`\\)" 1 font-lock-warning-face)
    ("^\s*[+-]?\\(add\\|remove\\|rename\\|change\\|set\\)" 1 font-lock-function-name-face)
    ("^\s*\\(name\\)" 1 font-lock-builtin-face)
    ("\\(\"[^\"]+\"\\)" 1 font-lock-warning-face)
    ("^\s*[+-]?\\(use\\|create\\|end\\)" 1 font-lock-builtin-face)
    ("\\(table\\|schema\\column\\|index\\|option\\|primary\\|engine\\)" 1 font-lock-keyword-face))
    '("\\.mj$")                              ;; files that trigger this mode
    nil                                      ;; any other functions to call
    "Mohajer mode, mohajer migration file"   ;; doc string
  )
