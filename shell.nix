with import <nixpkgs> {};

pkgs.mkShell {
  name = "X";
  buildInputs = [
    pkgs.grpc-tools
    pkgs.protobuf
    pkgs.protoc-gen-go
  ];

  shellHook = ''
      export GOPATH=$(${go}/bin/go env GOPATH)
      export GOROOT=$(${go}/bin/go env GOROOT)
      export GO111MODULE=on
  '';
}

