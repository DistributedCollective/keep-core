pragma solidity 0.5.17;

import "openzeppelin-solidity/contracts/token/ERC20/ERC20Burnable.sol";
import "openzeppelin-solidity/contracts/token/ERC20/SafeERC20.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "./TokenStaking.sol";
import "./utils/BytesLib.sol";
import "./KeepToken.sol";

contract TokenGrantStake {
    using SafeMath for uint256;
    using BytesLib for bytes;

    ERC20Burnable token;
    TokenStaking tokenStaking;

    address tokenGrant; // Address of the master grant contract.

    uint256 grantId; // ID of the grant for this stake.
    uint256 amount; // Amount of staked tokens.
    address operator; // Operator of the stake.

    constructor(
        address _tokenAddress,
        uint256 _grantId,
        address _tokenStaking
    ) public {
        require(
            _tokenAddress != address(0x0),
            "Token address can't be zero."
        );
        require(
            _tokenStaking != address(0x0),
            "Staking contract address can't be zero."
        );

        token = ERC20Burnable(_tokenAddress);
        tokenGrant = msg.sender;
        grantId = _grantId;
        tokenStaking = TokenStaking(_tokenStaking);
    }

    function stake(
        uint256 _amount,
        bytes memory _extraData
    ) public onlyGrant {
        amount = _amount;
        operator = _extraData.toAddress(20);
        tokenRecipient spender = tokenRecipient(address(tokenStaking));

        if (token.approve(address(tokenStaking), _amount)) {
            spender.receiveApproval(
                address(this),
                _amount,
                address(token),
                _extraData
            );
        }
    }

    function getGrantId() public view onlyGrant returns (uint256) {
        return grantId;
    }

    function getAmount() public view onlyGrant returns (uint256) {
        return amount;
    }

    function getStakingContract() public view onlyGrant returns (address) {
        return address(tokenStaking);
    }

    function getDetails() public view onlyGrant returns (
        uint256 _grantId,
        uint256 _amount,
        address _tokenStaking
    ) {
        return (
            grantId,
            amount,
            address(tokenStaking)
        );
    }

    function cancelStake() public onlyGrant returns (uint256) {
        tokenStaking.cancelStake(operator);
        return returnTokens();
    }

    function undelegate() public onlyGrant {
        tokenStaking.undelegate(operator);
    }

    function recoverStake() public onlyGrant returns (uint256) {
        tokenStaking.recoverStake(operator);
        return returnTokens();
    }

    function returnTokens() internal returns (uint256) {
        uint256 returnedAmount = token.balanceOf(address(this));
        amount -= returnedAmount;
        token.transfer(tokenGrant, returnedAmount);
        return returnedAmount;
    }

    modifier onlyGrant {
        require(
            msg.sender == tokenGrant,
            "For token grant contract only"
        );
        _;
    }
}
