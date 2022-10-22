#include <Python.h>

#include "libwasmd.h"
#include "wallet.h"

static PyMethodDef cwpyMethods[] = {
    {"init_wallet",  cwpy_init_wallet, METH_VARARGS, "Initializes a new wallet and returns the wallet object"},
    {NULL, NULL, 0, NULL}        /* Sentinel */
};

static struct PyModuleDef cwpyModule = {
    PyModuleDef_HEAD_INIT,
    "cwpy",
    NULL,
    -1,
    cwpyMethods
};

PyMODINIT_FUNC
PyInit_cwpy(void)
{
    PyObject *m;

    m = PyModule_Create(&cwpyModule);
    if (m == NULL) {
        return NULL;
    }

    // initialize cfg for wasmd
    cfgInit();

    return m;
}