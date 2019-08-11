// http://172.104.241.88:3000/dashboard

const WebSocket = require('ws')

const url = 'ws://172.104.241.88:8080'
//const url = 'ws://localhost:8080'
const connection = new WebSocket(url)

var messages = []
messages.push({ code: "PLACE-ORDER" , account: "DU997901" , op: "BUY" , symbol: "CSCO" , qty: 1 , order_type: "LMT" , price: 123.32})
messages.push({ code: "PLACE-ORDER" , account: "DU1031917" , op: "SELL" , symbol: "TEVA" , qty: 2 , order_type: "MKT" })
messages.push({ code: "PLACE-ORDER" , account: "DU997900" , op: "BUY" , symbol: "MSFT" , qty: 3 , order_type: "MKT" })

// ADDED MESSAGES
messages.push({ code: "ORDER-STATUS-REQUEST" , oid: 430 })
messages.push({ code: "CANCEL-ORDER-REQUEST" , oid: 430 })
messages.push({ code: "ACCOUNT-INFO-REQUEST" , account: "DU997900" })


function processMessage(e){
    console.log(e.data)
    return 0;
}

function sendMessage(m){
    connection.send(JSON.stringify(m));
}

connection.onopen = () => {
    messages.forEach( sendMessage);
}

connection.onerror = (error) => {
    console.log(`WebSocket error: ${error}`)
}

connection.onmessage = (e) => {
    setTimeout(function(){processMessage(e)}, 10);
}