{
    description = "Generate OpenAPI docs from HTTP files";
    nixConfig = {
        bash-prompt = "(httptoswagger) -> ";
    };
    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
        flake-utils.url = "github:numtide/flake-utils";
    };
    outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem(system:
        let
            pkgs = nixpkgs.legacyPackages.${system};
        in {
            devShells.default = with pkgs; mkShell {
                hardeningDisable = [ "fortify" ];
                nativeBuildInputs = [ git ];
                buildInputs = [ go ];
                # shellHook = ''
                # '';
            };
        }
    );
}
