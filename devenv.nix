{ pkgs, lib, ... }:
{
  # To resolve error with double registration of cachix
  cachix.enable = false;

  packages = [
    pkgs.pre-commit
    pkgs.go_1_24
    pkgs.golangci-lint
    pkgs.mdbook
    pkgs.graphviz
    pkgs.shellcheck
    pkgs.mdbook-yml-header
    pkgs.mdbook-variables
  ];
}
