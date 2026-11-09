{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:

{
  # Environment variables
  env = {
    PORT = "8080";
  };

  # Languages
  languages.go = {
    enable = true;
    package = pkgs.go_1_27;
  };

  # Packages
  packages = with pkgs; [
    git
    templ
    tailwindcss_4
    air
    sqlite

    # Language Servers (LSP)
    gopls
    tailwindcss-language-server
    nil
  ];

  # Development scripts
  scripts = {
    # Generate Templ templates
    "gen:templ".exec = "templ generate";

    # Build Tailwind CSS
    "build:css".exec = "tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify";

    # Watch Tailwind CSS
    "watch:css".exec = "tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --watch";

    # Full production build
    "build".exec = ''
      templ generate
      tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify
      go build -o bin/server main.go
    '';

    # Run dev server with hot reload
    "dev".exec = "air";
  };

  # Background processes managed by devenv up
  processes = {
    web.exec = "air";
  };

  enterShell = ''
    echo "=========================================================="
    echo "  Юридический сайт-визитка (Go + PocketBase + Templ + HTMX) "
    echo "=========================================================="
    echo "  Go:         $(go version)"
    echo "  Templ:      $(templ --version 2>/dev/null || echo 'доступен')"
    echo "  Tailwind:   $(tailwindcss --help 2>&1 | head -n 1)"
    echo "  Air:        $(air -v 2>/dev/null || echo 'доступен')"
    echo "  LSP:        gopls, templ lsp, tailwindcss-language-server, nil"
    echo ""
    echo "  Команды:"
    echo "    devenv up        - запуск всех процессов разработки"
    echo "    dev              - запуск air (live-reload)"
    echo "    build            - полная сборка проекта"
    echo "=========================================================="
  '';
}
