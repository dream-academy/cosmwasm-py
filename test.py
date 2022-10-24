import cwpy
import requests
import yaml, json

def get_umlg_from_faucet(addr):
    data = { "denom": "umlg", "address": addr }
    r = requests.post("https://faucet.malaga-420.cosmwasm.com/credit", json=data)
    if "OK" in r.text:
        return True
    else:
        return False

def code_id_from_log(log):
    logs = json.loads(yaml.load(log, Loader=yaml.FullLoader)['raw_log'])[0]
    code_id = logs["events"][-1]["attributes"][0]["value"]
    return int(code_id)

def store():
    with open("tests/phonebook.wasm", "rb") as f:
        wasmBytes = f.read()
    res = w.tx_wasm_store("procfs", wasmBytes)
    code_id = code_id_from_log(res)
    return code_id

def instantiate(code_id):
    imsg = b"{}"
    res = w.tx_wasm_instantiate("procfs", code_id, "label", imsg, 100)
    logs = json.loads(yaml.load(res, Loader=yaml.FullLoader)['raw_log'])[0]
    for e in logs['events']:
        if e["type"] == "instantiate":
            o = { x["key"]:x["value"] for x in e["attributes"]}
            assert(int(o["code_id"]) == code_id)
            return o["_contract_address"]

def execute(contract_addr):
    msg = json.dumps({"add_number":{"number":"010-9395-0000"}})
    res = w.tx_wasm_execute("procfs", contract_addr, msg, 100)
    print(res)

if __name__ == "__main__":
    w = cwpy.wallet("malaga-420", "https://rpc.malaga-420.cosmwasm.com:443")

    # initialize wallet with mnemonic
    mnemonics = "addict valve sudden antique budget honey cactus ten retreat over old admit swap summer shoe bachelor print glare ridge flower praise earn maze axis"
    w.add_key_with_mnemonic("procfs", mnemonics)
    addr = w.get_key("procfs")

    #store()
    #contract_addr = instantiate(1744)
    contract_addr = "wasm14k3zruqx7atjwzatga965dpyquksq2hy0ykfu6nq3g2fzqpv6pnsl3a9te"
    #execute(contract_addr)
    query_msg = json.dumps({"get_number": {"address": addr}})
    res = w.query_contract_smart(contract_addr, query_msg)
    print(res)