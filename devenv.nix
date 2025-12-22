{ pkgs, lib, ... }:
let
  mdbook-gitinfo = pkgs.rustPlatform.buildRustPackage rec {
    pname = "mdbook-gitinfo";
    version = "1.1.0";
    src = pkgs.fetchFromGitHub {
      owner = "compeng0001";
      repo  = pname;
      rev   = "v${version}";
      sha256 = "sha256-DCn1ArXsawezYRI6OaZ8JS+JiR2tO19Q+KW/oAuMmIk=";
    };
    cargoHash = "sha256-IB/I3rNebMLlyGZeryp6xfWGdPWuIJNbVKUyNzGmpGk=";
    doCheck = false;
  };
in
{
  packages = [
    pkgs.pre-commit
    pkgs.go_1_24
    pkgs.golangci-lint
    pkgs.mdbook
    pkgs.graphviz
    pkgs.shellcheck
    mdbook-gitinfo
  ];
}
