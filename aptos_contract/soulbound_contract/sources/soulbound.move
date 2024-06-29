module soulbound::soulbound {
    use std::string::{Self, String};
    use std::signer;
    use aptos_token_objects::aptos_token;
    use aptos_framework::object::{Self, Object}; // warning during compilation: for testing purpose only
    use aptos_token_objects::token; // warning during compilation: for testing purpose only
    

    const COLLECTION_NAME: vector<u8> = b"Unlimitedior";
    const COLLECTION_DESCRIPTION: vector<u8> = b"Description";
    const COLLECTION_URI: vector<u8> = b"example.com/uri";

    fun init_module(account: &signer) {
        let description = string::utf8(COLLECTION_DESCRIPTION);
        let name = string::utf8(COLLECTION_NAME);
        let uri = string::utf8(COLLECTION_URI);

        aptos_token::create_collection(
            account,
            description,
            100,
            name,
            uri,
            false,
            false,
            false,
            false,
            false,
            false,
            false,
            false,
            false,
            0, 
            1,
        );
    }

    public entry fun mint_soulbound_token(
        account: &signer,
        token_name: String,
        token_description: String,
        token_uri: String,
        soulbound_to: address
    ) {
        let name = string::utf8(COLLECTION_NAME);

        // Mint un token soulbound
        aptos_token::mint_soul_bound(
            account,
            name,
            token_description,
            token_name,
            token_uri,
            vector[],
            vector[],
            vector[],
            soulbound_to
        );
    }

    #[test(creator = @0x500, user1 = @0x250)]  
    fun test_mint(creator: &signer, user1: &signer) {
        init_module(creator);

        let token_name = string::utf8(b"Anima Token #1");
        let token_description = string::utf8(b"Anima Token #1 Description");
        let token_uri = string::utf8(b"Anima Token #1 URI/");
        let user1_addr = signer::address_of(user1);

        // Creator mints the Anima token for User1.
        mint_soulbound_token(
            creator,
            token_description,
            token_name,
            token_uri,
            user1_addr,
        );
    }

    #[test(creator = @0x500, user1 = @0x250)]
    #[expected_failure]
    fun test_transfer_should_fail(creator: &signer, user1: &signer) {
        init_module(creator);

        let token_name = string::utf8(b"Anima Token #1");
        let token_description = string::utf8(b"Anima Token #1 Description");
        let token_uri = string::utf8(b"Anima Token #1 URI/");
        let user1_addr = signer::address_of(user1);

        mint_soulbound_token(
            creator,
            token_description,
            token_name,
            token_uri,
            user1_addr,
        );

        let collection_name = string::utf8(COLLECTION_NAME);
        let token_address = token::create_token_address(
            &signer::address_of(creator),
            &collection_name,
            &token_name
        );
        let token = object::address_to_object<aptos_token::AptosToken>(token_address);

        // user1 transfers the token to creator. It should fail because the token is soulbound
        let creator_address = signer::address_of(creator);
        object::transfer(user1, token, creator_address);
    }
}