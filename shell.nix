{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "flares-dev";

  nativeBuildInputs = with pkgs; [
    go_1_26
    golangci-lint
    goreleaser
    gofumpt
    git
  ];
}
