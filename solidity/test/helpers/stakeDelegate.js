async function stakeDelegate(stakingContract, token, owner, operator, beneficiary, authorizer, stake) {
  let delegation = Buffer.concat([
    Buffer.from(beneficiary.substr(2), 'hex'),
    Buffer.from(operator.substr(2), 'hex'),
    Buffer.from(authorizer.substr(2), 'hex')
  ]);

  const success = await token.approve(stakingContract.address, stake, { from: owner })

  if (success) {
    return stakingContract.receiveApproval(owner, stake, token.address, delegation, { from: owner })
  } else {
    return false
  }
}

module.exports = stakeDelegate
