module payment::payment {

    use std::vector;
    use aptos_framework::coin; // warning during compilation: for testing purpose only
    use aptos_framework::aptos_coin;
    use aptos_framework::timestamp;
    use aptos_framework::event;
    use aptos_framework::signer;
    use aptos_framework::account; // warning during compilation: for testing purpose only
    use aptos_framework::aptos_account;

    struct ReceivedPayment has store, drop, copy {
        sender: address,
        amount: u64,
        timestamp: u64,
    }

    // Struct to store payments per address
    struct PaymentData has key, store {
        payments: vector<ReceivedPayment>,
    }

    #[event]
    struct ReceivedPaymentEvent has store, drop {
        sender: address,
        amount: u64,
        timestamp: u64,
    }

    // init_module is called when publishing contract (ie VM is calling it automatically)
    fun init_module(sender: &signer) {
          let initial_data = PaymentData {
            payments: vector::empty<ReceivedPayment>(),
        };
        move_to(sender, initial_data);
    }

    // Entry function to receive payments.
    // It transfers Aptos tokens from sender to the contract with the amount specified in the function call.
    // It logs the payment details in the ReceivedPaymentEvent event.
     public entry fun receive_payment(sender: &signer, amount: u64) acquires PaymentData {
        let sender_address = signer::address_of(sender); 
        let contract_address = @payment;  // @MODULE_NAME = Contract address

        // Using aptos_account::transfer_coins because it creates the recipient account if it does not exist
       aptos_account::transfer_coins<aptos_coin::AptosCoin>(sender, contract_address, amount);

        let timestamp = timestamp::now_seconds();

        let payment = ReceivedPayment {
            sender: sender_address,
            amount: amount,
            timestamp: timestamp,
        };
        
        // Register in global storage (because PaymentData has key decorator) if address is not registered
        if (!exists<PaymentData>(sender_address)) {
            let payments = vector::empty<ReceivedPayment>();
            let payment_data = PaymentData {
                payments: payments,
            };
            move_to(sender, payment_data);
        };

        // Borrow mut to change the vector in global storage for key sender_address
        let payment_data_ref = borrow_global_mut<PaymentData>(sender_address);
        vector::push_back(&mut payment_data_ref.payments, payment);

        // Logging the payment through an event
        event::emit(ReceivedPaymentEvent { sender: sender_address, amount, timestamp });
    }

    // View (not modyfing storage) function to get all received payments for a specific address
    #[view]
    public fun get_payments_for_address(address: address): vector<ReceivedPayment> acquires PaymentData {
        if (!exists<PaymentData>(address)) {
            return vector::empty<ReceivedPayment>()
        };
        let payment_data_ref = borrow_global<PaymentData>(address);
        payment_data_ref.payments
    }

    //------------------------------
    //          TESTS
    //------------------------------

    // helper function to setup the test environment
    // It creates accounts and mint coins for them
    #[test_only]
     public inline fun setup(
        aptos_framework: &signer,
        user: &signer,
        user2: &signer,
    ): (address, address) {
        timestamp::set_time_has_started_for_testing(aptos_framework);
        let (burn_cap, mint_cap) = aptos_coin::initialize_for_test(aptos_framework);

        let user_addr = signer::address_of(user);
        account::create_account_for_test(user_addr);
        coin::register<aptos_coin::AptosCoin>(user);

        let user2_addr = signer::address_of(user2);
        account::create_account_for_test(user2_addr);
        coin::register<aptos_coin::AptosCoin>(user2);

        let coins_user1 = coin::mint(10000, &mint_cap);
        coin::deposit(user_addr, coins_user1);
        let coins_user2 = coin::mint(20000, &mint_cap);
        coin::deposit(user2_addr, coins_user2);

        coin::destroy_burn_cap(burn_cap);
        coin::destroy_mint_cap(mint_cap);

        (user_addr, user2_addr)
    }

    //aptos_framework must be @0x1 so that aptos_coin::initialize_for_test(aptos_framework) works in helper function
    #[test(aptos_framework = @0x1, user = @0x111, user2 = @0x222)]
    fun test_receive_payment(
        aptos_framework: &signer,
        user: &signer,
        user2: &signer
        ) acquires PaymentData {
            let (user_addr, user2_addr) = setup(aptos_framework, user, user2);
            timestamp::set_time_has_started_for_testing(aptos_framework);

            init_module(user);

            let initial_balance = coin::balance<aptos_coin::AptosCoin>(user_addr);

            let payment_amount: u64 = 100;
            receive_payment(user, payment_amount);

            let final_balance = coin::balance<aptos_coin::AptosCoin>(user_addr);

            assert!(final_balance == initial_balance - payment_amount, 2);

            // Check the payment data
            let payments = get_payments_for_address(user_addr);
            assert!(vector::length(&payments) == 1, 3);
            let payment = *vector::borrow(&payments, 0);
            assert!(payment.amount == payment_amount, 4);

            // Check payment data for user2
            let payments = get_payments_for_address(user2_addr);
            assert!(vector::length(&payments) == 0, 6);

            // User2 calls the contract
            let initial_balance_user2 = coin::balance<aptos_coin::AptosCoin>(user2_addr);

            let payment_amount_user2: u64 = 1234;
            receive_payment(user2, payment_amount_user2);

            let final_balance_user2 = coin::balance<aptos_coin::AptosCoin>(user2_addr);

            assert!(final_balance_user2 == initial_balance_user2 - payment_amount_user2, 2);

            // Check contract balance
            let final_balance = coin::balance<aptos_coin::AptosCoin>(@payment);
            assert!(final_balance == payment_amount + payment_amount_user2, 5);
        }   
}
