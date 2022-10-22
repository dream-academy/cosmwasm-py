from distutils.core import setup, Extension

module = Extension('cwpy',
    sources = ['main.c', 'wallet.c'],
    include_dirs=["."],
    extra_objects=["libwasmd.a"],
    extra_link_args=["-pthread"])

setup (name = 'CosmWasm-Python',
       version = '1.0',
       description = 'Web3 like library for CosmWasm',
       ext_modules = [module])