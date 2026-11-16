# Writing acceptance criteria

Each acceptance criterion (AC) proves one business rule or one flow step.
Write it as Given / When / Then, so a tester can check it without asking.

## Shape

```
AC-<n> (<BR-x or flow step>): Given <start state>, when <one action>, then <observable result>.
```

- **Given**: the state before the action. Name the data that matters.
- **When**: one action by one actor.
- **Then**: what the user or another system can see. Use numbers.

## Good examples (an online shop)

- AC-1 (BR-1): Given a cart with items worth 49.00, when the customer opens
  checkout, then the delivery fee shows 5.00.
- AC-2 (BR-1): Given a cart with items worth 50.00, when the customer opens
  checkout, then the delivery fee shows 0.00.
- AC-3 (flow step 4): Given a saved cart from yesterday, when the customer
  signs in on another device, then the same items and quantities show.
- AC-4 (BR-3): Given a coupon that expired yesterday, when the customer
  applies it, then the message "This coupon has expired" shows and the total
  does not change.

## Bad examples and fixes

| Bad | Why | Better |
| --- | --- | --- |
| "Checkout is fast." | No number, no action | "Given a cart of 10 items, when the customer taps Pay, then the confirmation shows within 3 seconds." |
| "Coupons work correctly." | Not testable | One AC per coupon rule: valid, expired, already used, below minimum. |
| "Given a cart, when the customer pays and gets an email and the stock updates, then it works." | Many actions and results | Split into one AC per result. |
| "The system calls the payment service." | Describes design, not behavior | "Then the order shows status Paid." |

## Checks

- Every business rule has at least one AC.
- Every AC names its rule or flow step.
- Edge cases (empty, too large, expired, duplicate) have their own ACs.
