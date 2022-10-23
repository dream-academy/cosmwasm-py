from distutils.core import setup, Extension
import os

godir = os.path.join(os.environ["HOME"], "go")
libwasmvm_dir = os.path.join(godir, "pkg", "mod", "github.com", "!cosm!wasm", "wasmvm@v1.0.0", "api")

module = Extension('cwpy',
    sources = ['main.c', 'wallet.c'],
    include_dirs=["."],
    extra_objects=["libwasmd.a"],
    library_dirs=[libwasmvm_dir],
    runtime_library_dirs=[libwasmvm_dir],
    extra_link_args=["-pthread", "-lwasmvm.x86_64"])

setup (name = 'CosmWasm-Python',
       version = '1.0',
       description = 'Web3 like library for CosmWasm',
       ext_modules = [module])