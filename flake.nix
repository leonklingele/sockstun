{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs/nixos-24.11";
    nixpkgs-unstable.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        unstable = import nixpkgs-unstable { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = [
            pkgs.gnumake

            # Backend
            unstable.go
            unstable.golangci-lint
            pkgs.air
          ];

          shellHook = ''
            export GOPATH="$(realpath "$PWD")/.gopath"
            export GOBIN="$GOPATH/bin"
            export GOCACHE="$GOPATH/.cache"
            export PATH="$PATH:$GOBIN"
          '';
        };
      }
    );
}
