/*
  JHA 31/05/16

  (c) Eliad Technologies, Inc.
*/

package main

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
    "encoding/json"

  //"github.com/openblockchain/obc-peer/openchain/chaincode/shim"
    "github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

const userPrefix = "usr:"
const prizePrefix = "prz:"
const allUsersKey = "allUsers"
//const allPrizesKey = "allPrizes"

type User struct {
    UserId    string `json:"user_id"`
    Email     string `json:"email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    FbId      string `json:"fb_id"`
    Prizes    []Prize `json:"prizes"`
}

type Prize struct {
    UserId string `json:"user_id"`
    Label  string `json:"label"`
    Date   string `json:"date"`
}

/* ??
type AllPizes struct {
    Prizes  []Prize
}
*/

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    // Initialize the collection of commercial paper keys
    fmt.Println("Initializing kozyo")
    var blank []string
    blankBytes, _ := json.Marshal(&blank)
    if err := stub.PutState(allUsersKey, blankBytes); err != nil {
        fmt.Println("Failed to initialize kozyo")
    }
    /*
    if err := stub.PutState(allPrizesKey, blankBytes); err != nil {
        fmt.Println("Failed to initialize kozyo")
    }
    */

    fmt.Println("Initialization complete")
    return nil, nil
}

// Create a user
func (t *SimpleChaincode) createUser(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 5 {
        fmt.Printf("Error: createUser called with %d argument(s) (%v)\n",len(args),args)
        return nil, errors.New("createUser has 5 arguments")
    }

    userId := args[0]
    email  := args[1]
    firstName := args[2]
    lastName := args[3]
    fbId := args[4]
    user := User{UserId: userId, Email: email, FirstName: firstName, LastName: lastName, FbId: fbId }
    fmt.Printf("user = %T %v\n",user,user)
    newUserBytes, err := json.Marshal(&user)
    if err != nil  {
        msg := "Error marshalling " + userId
        fmt.Println(msg)
        return nil, errors.New(msg)
    }
    fmt.Printf("Marshall(user) -> %v\n",newUserBytes)

    userKey := userPrefix + userId
    fmt.Println("Attempting to get state for " + userKey)
    oldUserBytes, err := stub.GetState(userKey)
    if len(oldUserBytes) > 0 && err == nil {
        msg := fmt.Sprintf("Error: '%v' already exists (%T %v)",userKey,oldUserBytes,oldUserBytes)
        fmt.Println(msg)
        return nil, errors.New(msg)
    } else {
        fmt.Println("Put state "+userKey)
        err = stub.PutState(userKey, newUserBytes)
        if err != nil {
            fmt.Println("Error: put state " + userKey + " => " + err.Error())
            return nil, err
        }

        // Update users keys
        fmt.Println("Getting Users keys")

        keysBytes, err := stub.GetState(allUsersKey);
        if err != nil {
            fmt.Println("Error: retrieving " + allUsersKey + " => "+err.Error())
            return nil, err
        }
        fmt.Printf("GetState('%v') -> %v\n",allUsersKey,keysBytes)

        var keys []string
        err = json.Unmarshal(keysBytes, &keys)
        if err != nil {
            fmt.Println("Error: unmarshal " + allUsersKey + " => "+err.Error())
            return nil, err
        }
        fmt.Printf("Unmarshal(keysBytes) -> %v\n",keys)

        fmt.Println("Appending the new key to users keys")
        foundKey := false
        for _, key := range keys {
            if key == userKey {
                foundKey = true
            }
        }
        if foundKey == false {
            keys = append(keys, userKey)
            fmt.Printf("append(keys,'%v') -> %v\n",userKey,keys)

            keysBytesToWrite, err := json.Marshal(&keys)
            if err != nil {
                fmt.Println("Error: marshalling the keys => "+err.Error())
                return nil, err
            }
            fmt.Printf("Marshal(keys) -> %v\n",keysBytesToWrite)
            fmt.Println("Put state on "+allUsersKey)
            err = stub.PutState(allUsersKey, keysBytesToWrite)
            if err != nil {
                fmt.Println("Error: put state "+allUsersKey+" => "+err.Error())
                return nil, err
            }
        }

        fmt.Println("Created user " + userId)
        return nil, nil
    }
}

// ======================================================================================================
// Delete - remove a key/value pair from state
// ======================================================================================================
func (t *SimpleChaincode) delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 1 {
        fmt.Printf("Error: delete called with %d argument(s) (%v)\n",len(args),args)
        return nil, errors.New("delete has 1 arguments")
    }

    name := args[0]
    err := stub.DelState(name)                                                  //remove the key from chaincode state
    if err != nil {
        fmt.Println("Error: del state " + name + " => " + err.Error())
        return nil, err
    }
    fmt.Println("Del state '" + name + "' => OK")
    if strings.HasPrefix(name,"usr:") {
        keysBytes, err := stub.GetState(allUsersKey)
        if err != nil {
            fmt.Println("Error: retrieving " + allUsersKey + " => "+err.Error())
            return nil, err
        }
        fmt.Printf("GetState('%v') -> %v\n",allUsersKey,keysBytes)

        var keys []string
        err = json.Unmarshal(keysBytes, &keys)
        if err != nil {
            fmt.Println("Error: unmarshal " + allUsersKey + " => "+err.Error())
            return nil, err
        }
        fmt.Printf("Unmarshal(keysBytes) -> %v\n",keys)

        // Remove user from index
        for i,val := range keys {
            fmt.Println("keys["+strconv.Itoa(i) + "] '" + val + "' =?= '" + name+"'")
            if val == name {                                         // Find user
                fmt.Println("Found user at ["+strconv.Itoa(i) +"]")
                keys = append(keys[:i], keys[i+1:]...)               // Remove it
                for x:= range keys{                                  // Debug prints...
                    fmt.Println(string(x) + " - " + keys[x])
                }
                break
            }
        }
        jsonAsBytes, err := json.Marshal(keys)                                 //save new index
        if err != nil {
            fmt.Println("Error: marshalling the keys => "+err.Error())
            return nil, err
        }
        fmt.Println("Put state on "+allUsersKey)
        err = stub.PutState(allUsersKey, jsonAsBytes)
        if err != nil {
            fmt.Println("Error: put state "+allUsersKey+" => "+err.Error())
            return nil, err
        }
    }
    return nil, nil
}

// Run callback representing the invocation of a chaincode
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Println("run is running " + function)
    return t.Invoke(stub, function, args)
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Println("invoke is running " + function)
    // Handle different functions
    if function == "init" {                         // Initialize the chaincode
        return t.Init(stub, function, args)
    } else if function == "createUser" {            // Create a user
        return t.createUser(stub, args)
    } else if function == "delete" {                // Remove args[0] from state
        return t.delete(stub, args)
    }

    return nil, errors.New("Received unknown function '"+function+"' invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    if len(args) < 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting 1 or more argumenst")
    }
    if args[0] == "getAllUsers" {
        fmt.Println("Getting all users")
        allUsers, err := getAllUsers(stub)
        if err != nil {
            fmt.Println("Error from getAllUsers")   // Note: Error had printed in getAllUsers
            return nil, err
        } else {
            allUsersBytes, err1 := json.Marshal(&allUsers)
            if err1 != nil {
                fmt.Println("Error marshalling allUsers => "+err.Error())
                return nil, err1
            }
            fmt.Printf("Returning -> %v\n",allUsersBytes)
            return allUsersBytes, nil
        }
    } else {
        key := args[0]
        fmt.Println("Generic Query call, get state '"+key+"'")
        bytes, err := stub.GetState(key)
        if err != nil {
            fmt.Println("Error:  get state '"+key+"' => "+err.Error())
            return nil, err
        }

        fmt.Printf("Returning '%v' -> %v\n",key,bytes)
        return bytes, nil
    }
}

func getAllUsers(stub *shim.ChaincodeStub) ([]User, error){
    var allUsers []User
    // Get list of all the keys
    keysBytes, err := stub.GetState(allUsersKey)
    if err != nil {
        fmt.Println("Error get state "+allUsersKey+" => "+err.Error())
        return nil, err
    }
    fmt.Printf("GetState('%v') -> %v\n",allUsersKey,keysBytes)

    var keys []string
    err = json.Unmarshal(keysBytes, &keys)
    if err != nil {
        fmt.Println("Error unmarshalling "+allUsersKey+" => "+err.Error())
        return nil, err
    }
    fmt.Printf("Unmarshal(keysBytes) -> %v\n",keys)

    // Get all the Users
    for _, value := range keys {
        userBytes, err := stub.GetState(value)
        if err != nil {
            fmt.Println("Error: get state " + value+" => "+err.Error())
            return nil, err
        }
        fmt.Printf("GetState('%v') -> %v\n",value,userBytes)

        var user User
        err = json.Unmarshal(userBytes, &user)
        if err != nil {
            fmt.Println("Error: unmarshal " + value+" => "+err.Error())
            return nil, err
        }
        fmt.Printf("Unmarshal(userBytes) -> %v\n",user)

        fmt.Println("Appending " + value)
        allUsers = append(allUsers, user)
    }

    return allUsers, nil
}

func main() {
    err := shim.Start(new(SimpleChaincode))
    if err != nil {
        fmt.Printf("Error starting Simple chaincode: %s", err)
    }
}

