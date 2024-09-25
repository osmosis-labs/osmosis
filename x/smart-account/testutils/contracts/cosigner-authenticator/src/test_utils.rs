use cosmwasm_std::Binary;
use secp256k1::hashes::sha256;
use secp256k1::{Message, Secp256k1};
use secp256k1::{PublicKey, SecretKey};

/// Generates a Secp256k1 private/public key pair.
fn generate_keypair(hex_secret: &[u8]) -> (SecretKey, PublicKey) {
    let secp = Secp256k1::new();
    //let secret_key = SecretKey::from_str(randomness).unwrap();
    let secret_key = SecretKey::from_slice(&hex_secret).unwrap();
    let public_key = PublicKey::from_secret_key(&secp, &secret_key);
    (secret_key, public_key)
}

/// Signs a message with a given private key.
pub fn sign_message(priv_key: &SecretKey, message: &str) -> Vec<u8> {
    let secp = Secp256k1::new();
    let message = Message::from_hashed_data::<sha256::Hash>(message.as_bytes());
    //let message = Message::from_slice(message).expect("32 bytes");
    let signature = secp.sign_ecdsa(&message, priv_key);
    signature.serialize_compact().to_vec()
}

pub fn generate_keys_and_sign(hex_secret: &[u8], message: &str) -> (SecretKey, Binary, Binary) {
    let (priv_key, pub_key) = generate_keypair(hex_secret);
    let signature = sign_message(&priv_key, message);
    let pubkey_binary = Binary::from(pub_key.serialize().as_ref());
    let signature_binary = Binary::from(signature.as_slice());

    (priv_key, pubkey_binary, signature_binary)
}
