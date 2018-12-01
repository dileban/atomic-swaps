var GoldToken = artifacts.require("GoldToken");
var AtomicSwap = artifacts.require("AtomicSwap");

const secret = "secret"
const image = "0x2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b"

contract("AtomicSwap", (accounts) => {
    beforeEach(async () => {
        token = await GoldToken.new({from: accounts[0]});
        swap = await AtomicSwap.new({from: accounts[0]});
    });

    // Initial token balance is assigned to accounts[0]
    it("should allocate 21000000 tokens to the deployer's account", async () => {
        const balance = await token.balanceOf.call(accounts[0]);
        assert.strictEqual(balance.toNumber(), 21000000);
    });    
    
    // Locking tokens belonging to accounts[0] should increase AtomicSwap's balance
    it("should increase AtomicSwap's balance when locking 1500 tokens", async () => {
        await token.approve(swap.address, 2000);
        const allowance = await token.allowance.call(accounts[0], swap.address);
        assert.strictEqual(allowance.toNumber(), 2000);

        let balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 0);
        
        const lockReceipt = await swap.lock(accounts[1], image, 1500, token.address, 100);
        balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 1500);      
    });

    // Unlock tokens previously locked by accounts[0] after lock time has elapsed
    /*
    it("should unlock tokens locked by accounts[0] after lock time has elapsed", async () => {
        await token.approve(swap.address, 2000);
        const allowance = await token.allowance.call(accounts[0], swap.address);
        assert.strictEqual(allowance.toNumber(), 2000);

        const balance1 = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance1.toNumber(), 0);
        
        const lockReceipt = await swap.lock(accounts[1], image, 500, token.address, 100);
        const balance2 = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance2.toNumber(), 500);

        const unlockReceipt = await swap.unlock(lockReceipt.logs[0].args.agreementID);
        const balance3 = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance3.toNumber(), 0);        
    });
    */

    // Claiming 1500 tokens locked by accounts[0] should increase accounts[1]'s balance by 1500
    it("should increase accounts[1]'s balance by 1500 after claiming from accounts[0]", async () => {
        await token.approve(swap.address, 1500);
        const allowance = await token.allowance.call(accounts[0], swap.address);
        assert.strictEqual(allowance.toNumber(), 1500);

        let balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 0);
        
        const lockReceipt = await swap.lock(accounts[1], image, 1500, token.address, 100);
        balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 1500);

        const claimReceipt = await swap.claim(lockReceipt.logs[0].args.agreementID, secret, { from: accounts[1] });
        balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 0);

        balance = await token.balanceOf.call(accounts[1]);
        assert.strictEqual(balance.toNumber(), 1500);      
    });

    // Claiming more tokens than allowed in the agreement should fail.
    // This could be done by attempting to call claim twice.
    it("should fail attempts to claim more tokens than agreed upon", async () => {
        await token.approve(swap.address, 3000);
        const allowance = await token.allowance.call(accounts[0], swap.address);
        assert.strictEqual(allowance.toNumber(), 3000);

        let balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 0);

        // Lock for accounts[1]
        const lockReceipt1 = await swap.lock(accounts[1], image, 1000, token.address, 100);
        balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 1000);

        // Lock for accounts[2], AtomicSwap will now have 2000 tokens
        const lockReceipt2 = await swap.lock(accounts[2], image, 1000, token.address, 100);
        balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 2000);
        let status = await swap.getStatus.call(lockReceipt1.logs[0].args.agreementID);
        assert.strictEqual(status.toNumber(), 0);

        // accounts[1] claims 1000 tokens first, AtomicSwap's balance should now be 1000 tokens
        const claimReceipt1 = await swap.claim(lockReceipt1.logs[0].args.agreementID, secret, { from: accounts[1] });
        balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 1000);
        status = await swap.getStatus.call(lockReceipt1.logs[0].args.agreementID);
        assert.strictEqual(status.toNumber(), 2);       
        balance = await token.balanceOf.call(accounts[1]);
        assert.strictEqual(balance.toNumber(), 1000);

        // accounts[1] attempts to claims 1000 tokens again, this should fail and balances should remain unchanged.
        await assertRevert(swap.claim(lockReceipt1.logs[0].args.agreementID, secret, { from: accounts[1] }));
        balance = await token.balanceOf.call(swap.address);
        assert.strictEqual(balance.toNumber(), 1000);

        balance = await token.balanceOf.call(accounts[1]);
        assert.strictEqual(balance.toNumber(), 1000);      
    });   
});

// assertRevert checks to see if promise results in a VM revert exception. 
async function assertRevert(promise) {
    try {
        await promise;
        throw null;
    }
    catch(error) {
        // NOTE: Truffle doesn't support retrieving require messages from contracts yet.
        assert(error.message.includes("revert"), "Expected to see VM revert message");
    }
};
