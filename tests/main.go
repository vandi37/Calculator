// Where i haven't made any tests.
//
// 1. `../cmd/` : cmd runs the program it shouldn't be testes.
//
// 2. `../configs/` : configuration data shouldn't be testes
//
// 3. `../frontend/` : frontend shouldn't be tested using go).
//
// 4. `../internal/agent/get/` : it's hard to create tests there. (perhaps i will add tests later).
//
// 5. `../internal/agent/module/` : module shouldn't be testes.
//
// 6. `../internal/agent/task/` : there is nothing to test. the code is too easy.
//
// 7. `../internal/application/` : this code runs the app. it shouldn't be tested.
//
// 8. `../internal/config/` : loading config could be tested, however it is stupid.
//
// 9. `../internal/ms/` : there is nothing to test. the code is too easy.
//
// 10. `../internal/transport/` : the functionality of this package is based on other packages, so it has not much different.
package main
