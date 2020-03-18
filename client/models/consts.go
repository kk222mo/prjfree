package models

const BLOCK_SIZE = 64
const CLIENTS_COUNT = 3
const STABILITY = 3 // Describes count of clients which would recieve blocks
const TASK_TIME = 5000

const KEY_LEN = 2048 // RSA Key Length
const CONN_COUNT = 5 // Sessions count

/*
  Protocol description:

  With commutator:
    {MESSAGE};TO;MSG            Send message to another client
    {MESSAGEFORALL};MSG         Send message to all comutator clients
    {CLIENTS}                   Get commutator client list

  Between clients:
    [STARTSESSION]              Public keys exchange
    [ENCMESSAGE]                Get encrypted MESSAGE
    [FIND];<topic name>         Get all blocks with selected topic name

*/
