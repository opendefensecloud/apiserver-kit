{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:

{
  # https://devenv.sh/packages/
  packages = [
    pkgs.gnumake
    pkgs.jq
  ];

  # https://devenv.sh/languages/
  languages.go.enable = true;
  languages.go.package = pkgs.go_1_25;

  git-hooks.hooks = {
    gofmt.enable = true;
    golangci-lint.enable = true;
  };
  # See full reference at https://devenv.sh/reference/options/

  difftastic.enable = true;
  delta.enable = true;

  enterShell = ''
    make mod
  '';
}
