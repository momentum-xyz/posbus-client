import os
import subprocess
import pathlib
import shutil

from hatchling.builders.hooks.plugin.interface import BuildHookInterface


class CustomBuildHook(BuildHookInterface):
    def initialize(self, version, build_data):
        if self.target_name not in ["wheel"]:  # TODO "sdist"]:
            return
        self.clean([version])
        go_version = subprocess.check_output(["go", "env", "GOVERSION"]).decode().strip()
        go_path = subprocess.check_output(["go", "env", "GOPATH"]).decode().strip()
        env = os.environ.copy()
        env["PATH"] = f"{go_path}/bin:{env['PATH']}"
        if go_version.startswith("go1.19"):
            subprocess.check_call([
                "go", "install", "golang.org/dl/go1.20@latest"
                ])
            subprocess.check_call([
                f"{go_path}/bin/go1.20", "download"
                ], env=env)
            go_path = subprocess.check_output(["go1.20", "env", "GOPATH"], env=env).decode().strip()
            subprocess.check_call(["ln", "-s", "-f", f"{go_path}/bin/go1.20", f"{go_path}/bin/go"])
            env = os.environ.copy()
            env["PATH"] = f"{go_path}/bin:{env['PATH']}"

        subprocess.check_call([
            "go", "install", "golang.org/x/tools/cmd/goimports@v0.10.0",
            ], env=env)
        subprocess.check_call([
            "go", "install", "github.com/go-python/gopy@v0.4.7",
            ], env=env)
        env["CGO_LDFLAGS_ALLOW"] = "-fwrapv"
        subprocess.check_call([
            "gopy", "build", "-name=pbc", "-rename", "-dynamic-link=true",
            "-output=odyssey_posbus_client",
            "github.com/momentum-xyz/posbus-client/pbc/compat",
            "github.com/momentum-xyz/ubercontroller/pkg/posbus",
            "github.com/momentum-xyz/ubercontroller/pkg/cmath"
            ], env=env)
        build_data["infer_tag"] = True
        build_data["pure_python"] = False
        build_data["artifacts"] = ["odyssey_posbus_client/*"]
        build_data["only-include"] = ["odyssey_posbus_client"]

    def clean(self, version):
        src_files = [".gitignore", "hatch_build.py", "pyproject.toml", "README.md"]
        p = pathlib.Path(__file__).parent
        for child in p.iterdir():
            if child.is_file() and child.name not in src_files:
                child.unlink()
            elif child.is_dir() and child.name not in ["dist"]:
                shutil.rmtree(child)

