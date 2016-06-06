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
const diplomaPrefix = "dpl:"
const allUsersKey = "allUsers"
const allDiplomasKey = "allDiplomas"

type User struct {
    UserId    string `json:"user_id"`
    Email     string `json:"email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    FbId      string `json:"fb_id"`
    Diplomas  []string `json:"diplomas"`
}

type Diploma struct {
    DiplomaId string `json:"diploma_id"`
    UserId string `json:"user_id"`
    Label  string `json:"label"`
    Date   string `json:"date"`
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    // Initialize the collection of commercial paper keys
    fmt.Println("Initializing kozyo")
    var blank []string
    blankBytes, _ := json.Marshal(&blank)
    if err := stub.PutState(allUsersKey, blankBytes); err != nil {
        fmt.Println("Failed to initialize '"+allUsersKey+"'")
    }
    if err := stub.PutState(allDiplomasKey, blankBytes); err != nil {
        fmt.Println("Failed to initialize '"+allDiplomasKey+"'")
    }

    fmt.Println("Initialization complete")
    return nil, nil
}

// Create a user
func (t *SimpleChaincode) createUser(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 5 {
        fmt.Printf("Error: createUser called with %d argument(s) (%v)\n",len(args),args)
        return nil, errors.New("createUser has 5 arguments")
    }

    // Create a User with args
    userId := args[0]
    email  := args[1]
    firstName := args[2]
    lastName := args[3]
    fbId := args[4]
    var diplomas []string   // Empty array of diplomaId
    user := User{UserId: userId, Email: email, FirstName: firstName, LastName: lastName, FbId: fbId, Diplomas: diplomas }
    userKey := userPrefix + userId
    fmt.Printf("user '%s' = %T %v\n",userKey,user,user)

    // Marshal the structure
    newUserBytes, err := json.Marshal(&user)
    if err != nil  {
        msg := "Error marshalling " + userKey
        fmt.Println(msg)
        return nil, errors.New(msg)
    }
    fmt.Printf("Marshall(user) -> %v\n",newUserBytes)

    // Check if the user already exists
    fmt.Println("Attempting to get state for " + userKey)
    oldUserBytes, err := stub.GetState(userKey)
    if len(oldUserBytes) > 0 && err == nil {
        msg := fmt.Sprintf("Error: user '%v' already exists (%T %v)",userKey,oldUserBytes,oldUserBytes)
        fmt.Println(msg)
        return nil, errors.New(msg)
    } else {
        fmt.Println("Put state "+userKey)
        err = stub.PutState(userKey, newUserBytes)
        if err != nil {
            fmt.Println("Error: put state " + userKey + " => " + err.Error())
            return nil, err
        }

        // Update allUsersKey
        err = appendToKeyArray(stub,allUsersKey,userKey)
        if (err != nil) {
            return nil,err
        }

        fmt.Println("Created user " + userKey)
        return nil, nil
    }
}

// Create a diploma
func (t *SimpleChaincode) createDiploma(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 4 {
        fmt.Printf("Error: createDiploma called with %d argument(s) (%v)\n",len(args),args)
        return nil, errors.New("createDiploma has 4 arguments")
    }

    // Create a Diploma with args
    diplomaId := args[0]
    userId := args[1]
    label := args[2]
    date := args[3]
    diploma := Diploma{DiplomaId: diplomaId, UserId: userId, Label: label, Date: date }
    diplomaKey := diplomaPrefix + diplomaId
    fmt.Printf("diploma '%s' = %T %v\n",diplomaKey,diploma,diploma)

    // Marshal the structure
    newDiplomaBytes, err := json.Marshal(&diploma)
    if err != nil  {
        msg := "Error marshalling " + diplomaKey
        fmt.Println(msg)
        return nil, errors.New(msg)
    }
    fmt.Printf("Marshall(diploma) -> %v\n",newDiplomaBytes)

    // Retrieve the user
    userKey := userPrefix + userId
    fmt.Println("Attempting to get state for " + userKey)
    userBytes, err := stub.GetState(userKey)
    if err != nil {
        fmt.Println("Error: retrieving " + userKey + " => "+err.Error())
        return nil, err
    }
    if len(userBytes) == 0 {
        msg := fmt.Sprintf("Error: user '%v' does not exist (%T %v)",userKey,userBytes,userBytes)
        fmt.Println(msg)
        return nil, errors.New(msg)
    }
    var user User
    err = json.Unmarshal(userBytes, &user)
    if err != nil {
        fmt.Println("Error: unmarshal " + userKey+" => "+err.Error())
        return nil, err
    }
    fmt.Printf("Unmarshal(userBytes) -> %v\n",user)

    // Check if the diploma already exists
    fmt.Println("Attempting to get state for " + diplomaKey)
    oldDiplomaBytes, err := stub.GetState(diplomaKey)
    if len(oldDiplomaBytes) > 0 && err == nil {
        msg := fmt.Sprintf("Error: diploma '%v' already exists (%T %v)",diplomaKey,oldDiplomaBytes,oldDiplomaBytes)
        fmt.Println(msg)
        return nil, errors.New(msg)
    } else {
        fmt.Println("Put state "+diplomaKey)
        err = stub.PutState(diplomaKey, newDiplomaBytes)
        if err != nil {
            fmt.Println("Error: put state " + diplomaKey + " => " + err.Error())
            return nil, err
        }

        // Update user.Diplomas
        fmt.Println("Appending '"+diplomaKey+"' to user.Diplomas")
        foundKey := false
        for _, key := range user.Diplomas {
            if key == diplomaKey {
                foundKey = true
                break
            }
        }
        // Note: Should always be false
        if foundKey == false {
            user.Diplomas = append(user.Diplomas, diplomaKey)
            fmt.Printf("append(keys,'%v') -> %v\n",diplomaKey,user.Diplomas)

            // Marshal the structure
            newUserBytes, err := json.Marshal(&user)
            if err != nil  {
                msg := "Error marshalling " + userKey
                fmt.Println(msg)
                return nil, errors.New(msg)
            }
            fmt.Printf("Marshall(user) -> %v\n",newUserBytes)

            // Save user
            fmt.Println("Put state "+userKey)
            err = stub.PutState(userKey, newUserBytes)
            if err != nil {
                fmt.Println("Error: put state " + userKey + " => " + err.Error())
                return nil, err
            }
        }

        // Update allDiplomasKey
        err = appendToKeyArray(stub,allDiplomasKey,diplomaKey)
        if (err != nil) {
            return nil,err
        }

        fmt.Println("Created diploma " + diplomaKey)
        return nil, nil
    }
}

func appendToKeyArray(stub *shim.ChaincodeStub, arrayKey string,newKey string) error {
    fmt.Println("Get state '"+arrayKey+"'")
    keysBytes, err := stub.GetState(arrayKey);
    if err != nil {
        fmt.Println("Error: get state '" + arrayKey + "' => "+err.Error())
        return err
    }
    fmt.Printf("GetState('%v') -> %v\n",arrayKey,keysBytes)

    var keys []string
    err = json.Unmarshal(keysBytes, &keys)
    if err != nil {
        fmt.Println("Error: unmarshal " + arrayKey + " => "+err.Error())
        return err
    }
    fmt.Printf("Unmarshal(keysBytes) -> %v\n",keys)

    fmt.Println("Appending '"+newKey+"' to "+arrayKey+"[]")
    foundKey := false
    for _, key := range keys {
        if key == newKey {
            foundKey = true
            break
        }
    }
    // Note: Should always be false
    if foundKey == false {
        keys = append(keys, newKey)
        fmt.Printf("append(keys,'%v') -> %v\n",newKey,keys)

        keysBytesToWrite, err := json.Marshal(&keys)
        if err != nil {
            fmt.Println("Error: marshalling the keys => "+err.Error())
            return err
        }
        fmt.Printf("Marshal(keys) -> %v\n",keysBytesToWrite)
        fmt.Println("Put state '"+arrayKey+"'")
        err = stub.PutState(arrayKey, keysBytesToWrite)
        if err != nil {
            fmt.Println("Error: put state "+arrayKey+" => "+err.Error())
            return err
        }
    }
    return nil
}

func removeFromKeyArray(stub *shim.ChaincodeStub, arrayKey string,myKey string) error {
    keysBytes, err := stub.GetState(arrayKey)
    if err != nil {
        fmt.Println("Error: retrieving " + arrayKey + " => "+err.Error())
        return err
    }
    fmt.Printf("GetState('%v') -> %v\n",arrayKey,keysBytes)

    var keys []string
    err = json.Unmarshal(keysBytes, &keys)
    if err != nil {
        fmt.Println("Error: unmarshal " + arrayKey + " => "+err.Error())
        return err
    }
    fmt.Printf("Unmarshal(keysBytes) -> %v\n",keys)

    // Remove key from index
    for i,val := range keys {
        fmt.Println("keys["+strconv.Itoa(i) + "] '" + val + "' =?= '" + myKey+"'")
        if val == myKey {                                         // Find key
            fmt.Println("Found key at ["+strconv.Itoa(i) +"]")
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
        return err
    }
    fmt.Println("Put state on "+arrayKey)
    err = stub.PutState(arrayKey, jsonAsBytes)
    if err != nil {
        fmt.Println("Error: put state "+arrayKey+" => "+err.Error())
        return err
    }
    return nil;
}
// ======================================================================================================
// Delete - remove a key/value pair from state
// ======================================================================================================
func (t *SimpleChaincode) delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 1 {
        fmt.Printf("Error: delete called with %d argument(s) (%v)\n",len(args),args)
        return nil, errors.New("delete has 1 arguments")
    }

    key := args[0]
    // Check for additional clean-up to avoid references to deleted keys
    if strings.HasPrefix(key,userPrefix) {
        if err := userCleanup(stub,key); err != nil {
            return nil,err
        }
    } else if strings.HasPrefix(key,diplomaPrefix) {
        if err := diplomaCleanup(stub,key,true); err != nil {
            return nil,err
        }
    }

    err := stub.DelState(key)    // Remove the key from chaincode state
    if err != nil {
        fmt.Println("Error: del state " + key + " => " + err.Error())
        return nil, err
    }
    fmt.Println("Del state '" + key + "' => OK")
    return nil, nil
}

// Do additional clean-up when a user is deleted
func userCleanup(stub *shim.ChaincodeStub, userKey string) error {
    userBytes, err := stub.GetState(userKey)
    if err != nil {
        fmt.Println("Error: get state " + userKey+" => "+err.Error())
        return err
    }
    fmt.Printf("GetState('%v') -> %v\n",userKey,userBytes)

    if len(userBytes) > 0 {
        var user User
        err = json.Unmarshal(userBytes, &user)
        if err != nil {
            fmt.Println("Error: unmarshal " + userKey+" => "+err.Error())
            return err
        }
        fmt.Printf("Unmarshal(userBytes) -> %v\n",user)

        // Remove user's dipomas
        for _,diplomaKey := range user.Diplomas {
            if err := diplomaCleanup(stub,diplomaKey,false); err != nil {
                return err
            }
        }

        // Remove from users list
        if err := removeFromKeyArray(stub,allUsersKey,userKey); err != nil {
            return err
        }
    }

    return nil
}

// Do additional clean-up when a diploma is deleted
func diplomaCleanup(stub *shim.ChaincodeStub, diplomaKey string,doUserUpd bool) error {
    diplomaBytes, err := stub.GetState(diplomaKey)
    if err != nil {
        fmt.Println("Error: get state " + diplomaKey+" => "+err.Error())
        return err
    }
    fmt.Printf("GetState('%v') -> %v\n",diplomaKey,diplomaBytes)

    if len(diplomaBytes) > 0 {
        var diploma Diploma
        err = json.Unmarshal(diplomaBytes, &diploma)
        if err != nil {
            fmt.Println("Error: unmarshal " + diplomaKey+" => "+err.Error())
            return err
        }
        fmt.Printf("Unmarshal(diplomaBytes) -> %v\n",diploma)

        if doUserUpd {
            // Retrieve the user
            userKey := userPrefix+diploma.UserId
            userBytes, err := stub.GetState(userKey)
            if err != nil {
                fmt.Println("Error: get state " + userKey+" => "+err.Error())
                return err
            }
            fmt.Printf("GetState('%v') -> %v\n",userKey,userBytes)

            if len(userBytes) > 0 {
                var user User
                err = json.Unmarshal(userBytes, &user)
                if err != nil {
                    fmt.Println("Error: unmarshal " + userKey+" => "+err.Error())
                    return err
                }
                fmt.Printf("Unmarshal(userBytes) -> %v\n",user)

                // Remove from user's diplomas
                for i,val := range user.Diplomas {
                    fmt.Println("user.Diplomas["+strconv.Itoa(i) + "] '" + val + "' =?= '" + diplomaKey+"'")
                    if val == diplomaKey {                                                // Find key
                        fmt.Println("Found key at ["+strconv.Itoa(i) +"]")
                        user.Diplomas = append(user.Diplomas[:i], user.Diplomas[i+1:]...) // Remove it
                        for x:= range user.Diplomas{                                      // Debug prints...
                            fmt.Println(string(x) + " - " + user.Diplomas[x])
                        }

                        // Marshal the structure
                        newUserBytes, err := json.Marshal(&user)
                        if err != nil  {
                            msg := "Error marshalling " + userKey
                            fmt.Println(msg)
                            return errors.New(msg)
                        }
                        fmt.Printf("Marshall(user) -> %v\n",newUserBytes)

                        // Save user
                        fmt.Println("Put state "+userKey)
                        err = stub.PutState(userKey, newUserBytes)
                        if err != nil {
                            fmt.Println("Error: put state " + userKey + " => " + err.Error())
                            return err
                        }
                        break
                    }
                }
            }
        } else {
            // Remove from state
            if err := stub.DelState(diplomaKey); err != nil {
                fmt.Println("Error: del state " + diplomaKey + " => " + err.Error())
                return err
            }
        }

        // Remove from diplomas list
        if err := removeFromKeyArray(stub,allDiplomasKey,diplomaKey); err != nil {
            return err
        }
    }
    return nil
}

// Run callback representing the invocation of a chaincode
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Printf("Run(...,'%s',%v)\n",function,args)
    return t.Invoke(stub, function, args)
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Printf("Invoke(...,'%s',%v)\n",function,args)
    // Handle different functions
    if function == "init" {                         // Initialize the chaincode
        return t.Init(stub, function, args)
    } else if function == "createUser" {            // Create a user
        return t.createUser(stub, args)
    } else if function == "createDiploma" {            // Create a user
        return t.createDiploma(stub, args)
    } else if function == "delete" {                // Remove args[0] from state
        return t.delete(stub, args)
    }

    return nil, errors.New("Received unknown function '"+function+"' invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Printf("Query(...,'%s',%v)\n",function,args)
    if len(args) < 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting 1 or more argumenst")
    }
    if args[0] == "getAllUsers" {
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

func getAllUsers(stub *shim.ChaincodeStub) ([]User, error) {
    fmt.Println("getAllUsers()")
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
    for _, userKey := range keys {
        userBytes, err := stub.GetState(userKey)
        if err != nil {
            fmt.Println("Error: get state " + userKey+" => "+err.Error())
            return nil, err
        }
        fmt.Printf("GetState('%v') -> %v\n",userKey,userBytes)

        var user User
        err = json.Unmarshal(userBytes, &user)
        if err != nil {
            fmt.Println("Error: unmarshal " + userKey+" => "+err.Error())
            return nil, err
        }
        fmt.Printf("Unmarshal(userBytes) -> %v\n",user)

        // XXX JHA : ? convertir les clefs de user.Diplomas en structures ?

        fmt.Println("Appending " + userKey)
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

