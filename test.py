import cwpy

w = cwpy.wallet("malaga-420", "https://rpc.malaga-420.cosmwasm.com:443")
mnemonics = w.add_key_random("procfs")
print("mnemonics: {}".format(mnemonics))
addr = w.get_key("procfs")
print("address: {}".format(addr))