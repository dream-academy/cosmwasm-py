#include "wallet.h"
#include "libwasmd.h"
#include <stdint.h>

#define THROW_TYPE_ERROR(s) do {\
    PyErr_SetString(PyExc_TypeError, s);\
    return NULL;\
} while (0)

#define THROW_RUNTIME_ERROR(s) do {\
    PyErr_SetString(PyExc_RuntimeError, s);\
    return NULL;\
} while (0)

#define TYPECHECK_WALLET(wallet, maybe_wallet) do {\
    if (!PyObject_TypeCheck((maybe_wallet), &cwpyWalletType)) { \
        THROW_TYPE_ERROR("not wallet"); \
    } \
    wallet = (cwpyWallet *)(maybe_wallet); \
} while(0)

typedef struct cwpyWallet_s {
    PyObject_HEAD
    int walletId;
} cwpyWallet;

static PyObject* cwpyWalletAddKeyWithMnemonic(PyObject* self, PyObject* args);
static PyObject* cwpyWalletAddKeyRandom(PyObject* self, PyObject* args);
static PyObject* cwpyWalletGetKey(PyObject* self, PyObject* args);
static PyObject* cwpyWalletTxWasmStore(PyObject* self, PyObject* args);
static PyObject* cwpyWalletTxWasmInstantiate(PyObject* self, PyObject* args);
static PyObject* cwpyWalletTxWasmExecute(PyObject* self, PyObject* args);
static PyObject* cwpyWalletQueryContractStateSmart(PyObject* self, PyObject* args);
static PyObject* cwpyWalletQueryContractStateRaw(PyObject *self, PyObject *args);
static PyObject* cwpyWalletQueryContractStateAll(PyObject *self, PyObject *args);

static PyMethodDef cwpyWalletMethods[] = {
    {"add_key_random", cwpyWalletAddKeyRandom, METH_VARARGS, ""},
    {"add_key_with_mnemonic", cwpyWalletAddKeyWithMnemonic, METH_VARARGS, ""},
    {"get_key", cwpyWalletGetKey, METH_VARARGS, ""},
    {"tx_wasm_store", cwpyWalletTxWasmStore, METH_VARARGS, ""},
    {"tx_wasm_instantiate", cwpyWalletTxWasmInstantiate, METH_VARARGS, ""},
    {"tx_wasm_execute", cwpyWalletTxWasmExecute, METH_VARARGS, ""},
    {"query_contract_state_smart", cwpyWalletQueryContractStateSmart, METH_VARARGS, ""},
    {"query_contract_state_raw", cwpyWalletQueryContractStateRaw, METH_VARARGS, ""},
    {"query_contract_state_all", cwpyWalletQueryContractStateAll, METH_VARARGS, ""},
    { NULL, NULL, 0, NULL }
};

static PyObject *cwpyWalletNew(PyTypeObject *subtype, PyObject *args, PyObject *kwds) {
    char *chainId, *nodeUri;
    if (!PyArg_ParseTuple(args, "ss", &chainId, &nodeUri)) {
        THROW_TYPE_ERROR("argument types must be (str, str)");
    }
    cwpyWallet *self = (cwpyWallet *)subtype->tp_alloc(subtype, 0);
    if (self) {
        struct initWallet_return rv = initWallet(chainId, nodeUri);
        if (rv.r1 != NULL) {
            Py_DECREF(self);
            return Py_None;
        }
        self->walletId = rv.r0;
    }
    Py_INCREF(self);
    return (PyObject *)self;
}

static int cwpyWalletInit(PyObject *self, PyObject *args, PyObject *kwds) {
    return 0;
}

PyTypeObject cwpyWalletType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "wallet.wallet",
    .tp_doc = "",
    .tp_basicsize = sizeof(cwpyWallet),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = cwpyWalletNew,
    .tp_init = cwpyWalletInit,
    .tp_methods = cwpyWalletMethods,
};

static PyObject *cwpyWalletAddKeyRandom(PyObject *self, PyObject *args) {
    char *uid;
    cwpyWallet *wallet;
    struct addKeyRandom_return res;
    if (!PyArg_ParseTuple(args, "s", &uid)) {
        THROW_TYPE_ERROR("argument types must be (str)");
    }
    if (!PyObject_TypeCheck(self, &cwpyWalletType)) {
        THROW_TYPE_ERROR("not wallet");
    }
    wallet = (cwpyWallet *)self;
    res = addKeyRandom(wallet->walletId, uid);
    if (res.r1 != NULL) {
        THROW_RUNTIME_ERROR(res.r1);
    }
    PyObject *rv = PyUnicode_FromString(res.r0);
    Py_INCREF(rv);
    return rv;
}

static PyObject *cwpyWalletAddKeyWithMnemonic(PyObject *self, PyObject *args) {
    char *uid, *mnemonic;
    char *res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "ss", &uid, &mnemonic)) {
        THROW_TYPE_ERROR("argument types must be (str, str)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = addKeyMnemonic(wallet->walletId, uid, mnemonic);
    if (res != NULL) {
        THROW_RUNTIME_ERROR(res);
    }
    return Py_None;
}

static PyObject* cwpyWalletGetKey(PyObject* self, PyObject* args) {
    char *uid;
    struct getKey_return res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "s", &uid)) {
        THROW_TYPE_ERROR("argument types must be (str)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = getKey(wallet->walletId, uid);
    if (res.r1 != NULL) {
        return Py_None;
    }
    else {
        PyObject *rv = PyUnicode_FromString(res.r0);
        free(res.r0);
        Py_INCREF(rv);
        return rv;
    }
}

static PyObject* cwpyWalletTxWasmStore(PyObject* self, PyObject* args) {
    char *uid, *wasmData;
    Py_ssize_t wasmLen;
    struct txWasmStore_return res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "ss#", &uid, &wasmData, &wasmLen)) {
        THROW_TYPE_ERROR("argument types must be (str, bytes)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = txWasmStore(wallet->walletId, uid, wasmData, wasmLen);
    if (res.r1 != NULL) {
        THROW_RUNTIME_ERROR(res.r1);
    }
    else {
        PyObject *rv = PyUnicode_FromString(res.r0);
        free(res.r0);
        Py_INCREF(rv);
        return rv;
    }
}

static PyObject* cwpyWalletTxWasmInstantiate(PyObject* self, PyObject* args) {
    char *uid, *label, *msgData;
    Py_ssize_t msgLen;
    uint64_t codeId, umlgFunds;
    struct txWasmInstatitate_return res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "skss#k", &uid, &codeId, &label, &msgData, &msgLen, &umlgFunds)) {
        THROW_TYPE_ERROR("argument types must be (str, int, str, bytes, int)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = txWasmInstatitate(wallet->walletId, uid, codeId, label, msgData, msgLen, umlgFunds);
    if (res.r1 != NULL) {
        THROW_RUNTIME_ERROR(res.r1);
    }
    else {
        PyObject *rv = PyUnicode_FromString(res.r0);
        free(res.r0);
        Py_INCREF(rv);
        return rv;
    }
}

static PyObject* cwpyWalletTxWasmExecute(PyObject* self, PyObject* args) {
    char *uid, *contract, *msg;
    uint64_t umlgFunds;
    struct txWasmExecute_return res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "sssk", &uid, &contract, &msg, &umlgFunds)) {
        THROW_TYPE_ERROR("argument types must be (str, str, str, int)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = txWasmExecute(wallet->walletId, uid, contract, msg, umlgFunds);
    if (res.r1 != NULL) {
        THROW_RUNTIME_ERROR(res.r1);
    }
    else {
        PyObject *rv = PyUnicode_FromString(res.r0);
        free(res.r0);
        Py_INCREF(rv);
        return rv;
    }
}

static PyObject* cwpyWalletQueryContractStateSmart(PyObject* self, PyObject* args) {
    char *contract, *msg;
    struct queryContractStateSmart_return res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "ss", &contract, &msg)) {
        THROW_TYPE_ERROR("argument types must be (str, str)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = queryContractStateSmart(wallet->walletId, contract, msg);
    if (res.r1 != NULL) {
        THROW_RUNTIME_ERROR(res.r1);
    }
    else {
        PyObject *rv = PyUnicode_FromString(res.r0);
        free(res.r0);
        Py_INCREF(rv);
        return rv;
    }
}

static PyObject* cwpyWalletQueryContractStateRaw(PyObject *self, PyObject *args) {
    char *contract, *msg;
    struct queryContractStateRaw_return res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "ss", &contract, &msg)) {
        THROW_TYPE_ERROR("argument types must be (str, str)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = queryContractStateRaw(wallet->walletId, contract, msg);
    if (res.r1 != NULL) {
        THROW_RUNTIME_ERROR(res.r1);
    }
    else {
        PyObject *rv = PyUnicode_FromString(res.r0);
        free(res.r0);
        Py_INCREF(rv);
        return rv;
    }
}

static PyObject* cwpyWalletQueryContractStateAll(PyObject *self, PyObject *args) {
    char *contract;
    struct queryContractStateAll_return res;
    cwpyWallet *wallet;
    if (!PyArg_ParseTuple(args, "s", &contract)) {
        THROW_TYPE_ERROR("argument types must be (str)");
    }
    TYPECHECK_WALLET(wallet, self);
    res = queryContractStateAll(wallet->walletId, contract);
    if (res.r1 != NULL) {
        THROW_RUNTIME_ERROR(res.r1);
    }
    else {
        PyObject *rv = PyUnicode_FromString(res.r0);
        free(res.r0);
        Py_INCREF(rv);
        return rv;
    }
}