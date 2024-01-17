from locust import HttpUser, task

# Top 10 by volume pairs
# Some are currently unused but left here for reference.

# Currently used in tests
UOSMO         = "uosmo" 
ATOM          = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
stOSMO        = "ibc/D176154B0C63D1F9C6DCFB4F70349EBF2E2B5A87A05902F57A6AE92B863E9AEC"
stATOM        = "ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901"
# Currently used in tests
USDC          = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
USDCaxl       = "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"
USDT          = "ibc/4ABBEF4C8926DDDB320AE5188CFD63267ABBCEFC0583E4AE05D6E5AA2401DDAB"
WBTC          = "ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F"
ETH           = "ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5"
AKT           = "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4"
# Currently used in tests
UMEE          = "ibc/67795E528DF67C5606FC20F824EA39A6EF55BA133F4DC79C90A8C47A0901E17C"


top10ByVolumePairs = [
    UOSMO,
    ATOM,
    stOSMO,
    stATOM,
    USDC,
    USDCaxl,
    USDT,
    WBTC,
    ETH,
    AKT,
    UMEE
];


class SQS(HttpUser):

    # all-pools endpoint

    @task
    def all_pools(self):
        self.client.get("/all-pools")
    
    # Quote the same pair of UOSMO and USDC (UOSMO in) while progressively
    # increasing the amount of the tokenIn per endpoint.

    @task
    def quoteUOSMOUSDC_1In(self):
        self.client.get(f"/quote?tokenIn=1000000{UOSMO}&tokenOutDenom={USDC}")

    @task
    def quoteUOSMOUSDC_1000In(self):
        self.client.get(f"/quote?tokenIn=1000000000{UOSMO}&tokenOutDenom={USDC}")

    @task
    def quoteUOSMOUSDC_1000000In(self):
        self.client.get(f"/quote?tokenIn=1000000000000{UOSMO}&tokenOutDenom={USDC}")

    @task
    def singleQuoteUOSMOUSDC_1000000In(self):
        self.client.get(f"/single-quote?tokenIn=1000000000000{UOSMO}&tokenOutDenom={USDC}")

    # Quote the same pair of UOSMO and USDC (USDC in).
    @task
    def quoteUSDCUOSMO_1000000In(self):
        self.client.get(f"/quote?tokenIn=100000000000{USDC}&tokenOutDenom={UOSMO}")

    @task
    def quoteUSDCTUMEE_3000IN(self):
        self.client.get(f"/quote?tokenIn=3000000000{USDT}&tokenOutDenom={UMEE}")

    @task
    def routesUOSMOUSDC(self):
        self.client.get(f"/routes?tokenIn={UOSMO}&tokenOutDenom={USDC}")

    
    @task
    def routesUSDCUOSMO(self):
        self.client.get(f"/routes?tokenIn={USDC}&tokenOutDenom={UOSMO}")

    # TODO:
    # Add tests for routes search

