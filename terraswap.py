import cwpy
import requests
import yaml, json

def get_umlg_from_faucet(addr):
    data = { "denom": "umlg", "address": addr }
    r = requests.post("https://faucet.malaga-420.cosmwasm.com/credit", json=data)
    if "ok" in r.text:
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

def instantiate(code_id, imsg):
    if type(imsg) == str:
        imsg = imsg.encode()
    res = w.tx_wasm_instantiate("procfs", code_id, "label", imsg, 100)
    logs = json.loads(yaml.load(res, Loader=yaml.FullLoader)['raw_log'])[0]
    for e in logs['events']:
        if e["type"] == "instantiate":
            o = { x["key"]:x["value"] for x in e["attributes"]}
            assert(int(o["code_id"]) == code_id)
            return o["_contract_address"]
    raise Exception("code_id {} does not exist".format(code_id))

def execute(contract_addr, msg, value=100):
    res = w.tx_wasm_execute("procfs", contract_addr, msg, value)
    return res

if __name__ == "__main__":
    w = cwpy.wallet("malaga-420", "https://rpc.malaga-420.cosmwasm.com:443")

    # initialize wallet with mnemonic
    mnemonics = "addict valve sudden antique budget honey cactus ten retreat over old admit swap summer shoe bachelor print glare ridge flower praise earn maze axis"
    w.add_key_with_mnemonic("procfs", mnemonics)
    my_addr = w.get_key("procfs")

    '''
    with open("/home/procfs/terraswap/artifacts/terraswap_factory.wasm", "rb") as f:
        factory_code = f.read()
    res = w.tx_wasm_store("procfs", factory_code)
    factory_code_id = code_id_from_log(res)
    print("factory_code_id: {}".format(factory_code_id))

    with open("/home/procfs/terraswap/artifacts/terraswap_pair.wasm", "rb") as f:
        pair_code = f.read()
    res = w.tx_wasm_store("procfs", pair_code)
    pair_code_id = code_id_from_log(res)
    print("pair_code_id: {}".format(pair_code_id))

    with open("/home/procfs/terraswap/artifacts/terraswap_token.wasm", "rb") as f:
        token_code = f.read()
    res = w.tx_wasm_store("procfs", token_code)
    token_code_id = code_id_from_log(res)
    print("token_code_id: {}".format(token_code_id))

    imsg = json.dumps({"pair_code_id": 1786, "token_code_id": 1787})
    factory_addr = instantiate(1785, imsg)
    print("factory_addr: {}".format(factory_addr))

    msg = json.dumps({"add_native_token_decimals": {"denom": "umlg", "decimals": 6}})
    res = execute(factory_addr, msg)

    imsg = json.dumps({ "name": "DreamToken", "symbol": "DTC", "decimals": 6, "initial_balances": [ { "address": my_addr, "amount": "1000000" } ], "mint": { "minter": my_addr, "cap": "99900000000" } })
    token_addr = instantiate(token_code_id, imsg)
    print("token_addr: {}".format(token_addr))

    msg = json.dumps({"create_pair": {"asset_infos": [{"native_token": {"denom": "umlg"}}, {"token": {"contract_addr": token_addr}}]}})
    res = execute(factory_addr, msg)
    print(res)

    msg = json.dumps({"increase_allowance": {"spender": pair_addr, "amount": str(2**128-1), "expires": None}})
    res = execute(token_addr, msg)
    print(res)

    msg = json.dumps({"allowance": {"owner": my_addr, "spender": pair_addr}})
    res = w.query_contract_state_smart(token_addr, msg)
    print(res)

    msg = json.dumps({"provide_liquidity": {
        "assets": [
            {
                "info": { "token": { "contract_addr": token_addr } },
                "amount": "1000000"
            },
            {
                "info": { "native_token": { "denom": "umlg" } },
                "amount": "1000000"
            }
        ],
        "slippage_tolerance": None,
        "receiver": None
    }})
    res = execute(pair_addr, msg, 1000000)
    print(res)

    msg = json.dumps({"balance": {"address": my_addr}})
    res = w.query_contract_state_smart(lptoken_addr, msg)
    print(res)

    msg = json.dumps({
        "simulation": {
            "offer_asset": {
                "info": { "native_token": { "denom": "umlg" } },
                "amount": "100"
            }
        }
    })
    res = w.query_contract_state_smart(pair_addr, msg)
    print(res)

    msg = json.dumps({"swap":
    {
        "offer_asset": {
            "info": { "native_token": { "denom": "umlg" } },
            "amount": "100"
        },
        "belief_price": None,
        "max_spread": None,
        "to": None
    }})
    res = execute(pair_addr, msg)
    print(res)
    '''

    factory_code_id = 1785
    pair_code_id = 1786
    token_code_id = 1787
    factory_addr = "wasm1hczjykytm4suw4586j5v42qft60gc4j307gf7cxuazfg7jxt4h4sjvp7rx"
    token_addr = "wasm124v54ngky9wxhx87t252x4xfgujmdsu7uhjdugtkkqt39nld0e6st7e64h"
    pair_addr = "wasm15le5evw4regnwf9lrjnpakr2075fcyp4n4yzpelvqcuevzkw2lss46hslz"
    lptoken_addr = "wasm147ntaasx8mcx6a8jk7cvpyvus8r80garfnue4qrzrl0whk9ftntqpld03t"