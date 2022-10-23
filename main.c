#include <Python.h>

#include "libwasmd.h"
#include "wallet.h"

static PyMethodDef cwpyMethods[] = {
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

    libwasmdInit();

    // initialize type
    if (PyType_Ready(&cwpyWalletType) < 0) {
        return NULL;
    }
    Py_INCREF(&cwpyWalletType);
    PyModule_AddObject(m, "wallet", (PyObject *) &cwpyWalletType);

    return m;
}