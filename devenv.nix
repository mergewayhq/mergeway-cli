{ pkgs, lib, ... }:
{
  # To resolve error with double registration of cachix
  cachix.enable = false;

  languages.javascript = {
    enable = true;
    npm.enable = true;
  };

  packages = [
    pkgs.pre-commit
    pkgs.yamllint
    pkgs.go_1_26
    pkgs.golangci-lint
    pkgs.mdbook
    pkgs.graphviz
    pkgs.shellcheck
    pkgs.mdbook-yml-header
    pkgs.mdbook-variables
  ];
}
