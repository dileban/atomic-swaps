module.exports = {
    networks: {
        development: {
            host: "127.0.0.1",
            port: 7545,
            network_id: "*"
        },
        rinkebyInfura: {
            provider: function() {
            },
            network_id: "4"          
        },       
        rinkebyLocal: {
            host: "127.0.0.1",
            port: 8545,
            network_id: "4"
        }
    }
};
