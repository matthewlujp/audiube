package main

import (
	"fmt"
	"net/http"
	"sync"
)

// variableManager is for variables sharing among handlers
type variableManager struct {
	lock      sync.RWMutex
	variables map[*http.Request]map[string]interface{}
}

// vManager is a singleton instance of variableManager
var vManager = variableManager{
	variables: make(map[*http.Request]map[string]interface{}),
}

// register a map of variables for a request
// if a request has already been registered as a key, an error is returned
func (vm *variableManager) register(req *http.Request) error {
	vm.lock.Lock()
	defer vm.lock.Unlock()

	if _, ok := vm.variables[req]; ok {
		// req is already registered
		return fmt.Errorf("request %v exists in variables manager", *req)
	}

	// set to the map
	vm.variables[req] = make(map[string]interface{})
	return nil
}

// deregister delete entry for req from variables
// if a request does not exist, an error is returned
func (vm *variableManager) deregister(req *http.Request) error {
	vm.lock.Lock()
	defer vm.lock.Unlock()

	if _, ok := vm.variables[req]; !ok {
		// req does not exist
		return fmt.Errorf("request %v does not exist in variable manager", *req)
	}

	// remove entry
	delete(vm.variables, req)
	return nil
}

// get a variable designated by a req and a variable name
// if a target variable does not exist, return nil
func (vm *variableManager) get(req *http.Request, varName string) interface{} {
	vm.lock.RLock()
	defer vm.lock.RUnlock()

	varMap, _ := vm.variables[req]
	if varMap == nil { // map does not exist
		return nil
	}
	v, _ := varMap[varName]
	return v
}

// set variable with key req and variable name
// if a map for req does not exist, create it first
func (vm *variableManager) set(req *http.Request, varName string, variable interface{}) {
	vm.lock.Lock()
	defer vm.lock.Unlock()

	_, ok := vm.variables[req]
	if !ok { // req is not registered
		// create a map and set to variables
		vm.variables[req] = make(map[string]interface{})
	}

	// set variable
	vm.variables[req][varName] = variable
}

// withVars wraps http.HandlerFunc and register a map of variable for a request.
// other handlers can reference to shared variables for a request.
func withVars(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// register a map for variables which are related to a single request and shared among handlers
		if err := vManager.register(r); err != nil {
			http.Error(w, "failed to register variable, "+err.Error(), http.StatusInternalServerError)
		}
		f(w, r)
	}
}
