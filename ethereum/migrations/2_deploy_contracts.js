var AssetBackedToken = artifacts.require("AssetBackedToken");

module.exports = function(deployer) {
  deployer.deploy(AssetBackedToken);
};



