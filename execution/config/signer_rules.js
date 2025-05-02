// Automatically approve any transaction on chain ID 12345 (dev)
function ApproveTx(req) {
    if (req.chainId === 12345) {
        return "Approve";
    }
    return "Reject";
}

// Automatically approve message signing too
function ApproveSignData(req) {
    return "Approve";
}

// For all other requests
function Approve(req) {
    return "Approve";
}
