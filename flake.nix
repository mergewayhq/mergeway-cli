{
  description = "mergeway-cli: The official CLI for Mergeway";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "mergeway-cli";
          version = "0.1.0"; # Versioning can be improved later
          src = self;

          vendorHash = "sha256-pO4KEW2S84NepKekk1VMd+fG6pV7/DlPEwZgqgroyD0=";

          ldflags =
            let
              # self.rev is only available when the tree is clean and committed.
              # Use "dirty" if self.rev is undefined (e.g. dirty working tree).
              commit = self.rev or self.dirtyRev;
              
              # Date format helper
              formatIso8601 = ts:
                "${builtins.substring 0 4 ts}-${builtins.substring 4 2 ts}-${builtins.substring 6 2 ts}T${builtins.substring 8 2 ts}:${builtins.substring 10 2 ts}:${builtins.substring 12 2 ts}Z";
              
              buildDate = formatIso8601 self.lastModifiedDate;
            in
            [
              "-X github.com/mergewayhq/mergeway-cli/internal/version.Commit=${commit}"
              "-X github.com/mergewayhq/mergeway-cli/internal/version.BuildDate=${buildDate}"
            ];

          subPackages = [ "." ];

          meta = with pkgs.lib; {
            description = "The official CLI for Mergeway";
            homepage = "https://github.com/mergewayhq/mergeway-cli";
            license = licenses.mit;
            mainProgram = "mergeway-cli";
          };
        };
      }
    );
}
