var GoldToken = artifacts.require("GoldToken");
var AtomicSwap = artifacts.require("AtomicSwap");

module.exports = function(deployer) {
	 deployer.deploy(GoldToken);
	 deployer.deploy(AtomicSwap);
};



