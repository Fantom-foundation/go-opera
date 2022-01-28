// SPDX-License-Identifier: AGPL-3.0-only
pragma solidity ^0.8.4;

import {ISynthereumFinder} from './common/core/interfaces/IFinder.sol';
import {ISynthereumDeployer} from './interfaces/IDeployer.sol';
import {
  ISynthereumFactoryVersioning
} from './interfaces/IFactoryVersioning.sol';
import {ISynthereumRegistry} from './registries/interfaces/IRegistry.sol';
import {ISynthereumManager} from './interfaces/IManager.sol';
import {IERC20} from './@openzeppelin/contracts/token/ERC20/IERC20.sol';
import {IDeploymentSignature} from './interfaces/IDeploymentSignature.sol';
import {ISynthereumDeployment} from './common/interfaces/IDeployment.sol';
import {
  IAccessControlEnumerable
} from './@openzeppelin/contracts/access/IAccessControlEnumerable.sol';
import {SynthereumInterfaces, FactoryInterfaces} from './Constants.sol';
import {Address} from './@openzeppelin/contracts/utils/Address.sol';
import {
  ReentrancyGuard
} from './@openzeppelin/contracts/security/ReentrancyGuard.sol';
import {
  AccessControlEnumerable
} from './@openzeppelin/contracts/access/AccessControlEnumerable.sol';

contract SynthereumDeployer is
  ISynthereumDeployer,
  AccessControlEnumerable,
  ReentrancyGuard
{
  using Address for address;

  bytes32 private constant ADMIN_ROLE = 0x00;

  bytes32 public constant MAINTAINER_ROLE = keccak256('Maintainer');

  bytes32 private constant MINTER_ROLE = keccak256('Minter');

  bytes32 private constant BURNER_ROLE = keccak256('Burner');

  //Describe role structure
  struct Roles {
    address admin;
    address maintainer;
  }

  //----------------------------------------
  // State variables
  //----------------------------------------

  ISynthereumFinder public immutable synthereumFinder;

  //----------------------------------------
  // Events
  //----------------------------------------

  event PoolDeployed(uint8 indexed poolVersion, address indexed newPool);

  event SelfMintingDerivativeDeployed(
    uint8 indexed selfMintingDerivativeVersion,
    address indexed selfMintingDerivative
  );

  event FixedRateDeployed(
    uint8 indexed fixedRateVersion,
    address indexed fixedRate
  );

  //----------------------------------------
  // Modifiers
  //----------------------------------------

  modifier onlyMaintainer() {
    require(
      hasRole(MAINTAINER_ROLE, msg.sender),
      'Sender must be the maintainer'
    );
    _;
  }

  //----------------------------------------
  // Constructor
  //----------------------------------------

  /**
   * @notice Constructs the SynthereumDeployer contract
   * @param _synthereumFinder Synthereum finder contract
   * @param roles Admin and Maintainer roles
   */
  constructor(ISynthereumFinder _synthereumFinder, Roles memory roles) {
    synthereumFinder = _synthereumFinder;
    _setRoleAdmin(DEFAULT_ADMIN_ROLE, DEFAULT_ADMIN_ROLE);
    _setRoleAdmin(MAINTAINER_ROLE, DEFAULT_ADMIN_ROLE);
    _setupRole(DEFAULT_ADMIN_ROLE, roles.admin);
    _setupRole(MAINTAINER_ROLE, roles.maintainer);
  }

  //----------------------------------------
  // External functions
  //----------------------------------------

  /**
   * @notice Deploy a new pool
   * @param poolVersion Version of the pool contract
   * @param poolParamsData Input params of pool constructor
   * @return pool Pool contract deployed
   */
  function deployPool(uint8 poolVersion, bytes calldata poolParamsData)
    external
    override
    onlyMaintainer
    nonReentrant
    returns (ISynthereumDeployment pool)
  {
    pool = _deployPool(getFactoryVersioning(), poolVersion, poolParamsData);
    checkDeployment(pool, poolVersion);
    setSyntheticTokenRoles(pool);
    ISynthereumRegistry poolRegistry = getPoolRegistry();
    poolRegistry.register(
      pool.syntheticTokenSymbol(),
      pool.collateralToken(),
      poolVersion,
      address(pool)
    );
    emit PoolDeployed(poolVersion, address(pool));
  }

  /**
   * @notice Deploy a new self minting derivative contract
   * @param selfMintingDerVersion Version of the self minting derivative contract
   * @param selfMintingDerParamsData Input params of self minting derivative constructor
   * @return selfMintingDerivative Self minting derivative contract deployed
   */
  function deploySelfMintingDerivative(
    uint8 selfMintingDerVersion,
    bytes calldata selfMintingDerParamsData
  )
    external
    override
    onlyMaintainer
    nonReentrant
    returns (ISynthereumDeployment selfMintingDerivative)
  {
    ISynthereumFactoryVersioning factoryVersioning = getFactoryVersioning();
    selfMintingDerivative = _deploySelfMintingDerivative(
      factoryVersioning,
      selfMintingDerVersion,
      selfMintingDerParamsData
    );
    checkDeployment(selfMintingDerivative, selfMintingDerVersion);
    address tokenCurrency = address(selfMintingDerivative.syntheticToken());
    addSyntheticTokenRoles(tokenCurrency, address(selfMintingDerivative));
    ISynthereumRegistry selfMintingRegistry = getSelfMintingRegistry();
    selfMintingRegistry.register(
      selfMintingDerivative.syntheticTokenSymbol(),
      selfMintingDerivative.collateralToken(),
      selfMintingDerVersion,
      address(selfMintingDerivative)
    );
    emit SelfMintingDerivativeDeployed(
      selfMintingDerVersion,
      address(selfMintingDerivative)
    );
  }

  /**
   * @notice Deploy a fixed rate wrapper
   * @param fixedRateVersion Version of the fixed rate wrapper contract
   * @param fixedRateParamsData Input params of the fixed rate wrapper constructor
   * @return fixedRate FixedRate wrapper deployed
   */

  function deployFixedRate(
    uint8 fixedRateVersion,
    bytes calldata fixedRateParamsData
  )
    external
    override
    onlyMaintainer
    nonReentrant
    returns (ISynthereumDeployment fixedRate)
  {
    fixedRate = _deployFixedRate(
      getFactoryVersioning(),
      fixedRateVersion,
      fixedRateParamsData
    );
    checkDeployment(fixedRate, fixedRateVersion);
    setSyntheticTokenRoles(fixedRate);
    ISynthereumRegistry fixedRateRegistry = getFixedRateRegistry();
    fixedRateRegistry.register(
      fixedRate.syntheticTokenSymbol(),
      fixedRate.collateralToken(),
      fixedRateVersion,
      address(fixedRate)
    );
    emit FixedRateDeployed(fixedRateVersion, address(fixedRate));
  }

  //----------------------------------------
  // Internal functions
  //----------------------------------------

  /**
   * @notice Deploys a pool contract of a particular version
   * @param factoryVersioning factory versioning contract
   * @param poolVersion Version of pool contract to deploy
   * @param poolParamsData Input parameters of constructor of the pool
   * @return pool Pool deployed
   */
  function _deployPool(
    ISynthereumFactoryVersioning factoryVersioning,
    uint8 poolVersion,
    bytes memory poolParamsData
  ) internal returns (ISynthereumDeployment pool) {
    address poolFactory =
      factoryVersioning.getFactoryVersion(
        FactoryInterfaces.PoolFactory,
        poolVersion
      );
    bytes memory poolDeploymentResult =
      poolFactory.functionCall(
        abi.encodePacked(getDeploymentSignature(poolFactory), poolParamsData),
        'Wrong pool deployment'
      );
    pool = ISynthereumDeployment(abi.decode(poolDeploymentResult, (address)));
  }

  /**
   * @notice Deploys a self minting derivative contract of a particular version
   * @param factoryVersioning factory versioning contract
   * @param selfMintingDerVersion Version of self minting derivate contract to deploy
   * @param selfMintingDerParamsData Input parameters of constructor of self minting derivative
   * @return selfMintingDerivative Self minting derivative deployed
   */
  function _deploySelfMintingDerivative(
    ISynthereumFactoryVersioning factoryVersioning,
    uint8 selfMintingDerVersion,
    bytes calldata selfMintingDerParamsData
  ) internal returns (ISynthereumDeployment selfMintingDerivative) {
    address selfMintingDerFactory =
      factoryVersioning.getFactoryVersion(
        FactoryInterfaces.SelfMintingFactory,
        selfMintingDerVersion
      );
    bytes memory selfMintingDerDeploymentResult =
      selfMintingDerFactory.functionCall(
        abi.encodePacked(
          getDeploymentSignature(selfMintingDerFactory),
          selfMintingDerParamsData
        ),
        'Wrong self-minting derivative deployment'
      );
    selfMintingDerivative = ISynthereumDeployment(
      abi.decode(selfMintingDerDeploymentResult, (address))
    );
  }

  /**
   * @notice Deploys a fixed rate wrapper contract of a particular version
   * @param factoryVersioning factory versioning contract
   * @param fixedRateVersion Version of the fixed rate wrapper contract to deploy
   * @param fixedRateParamsData Input parameters of constructor of the fixed rate wrapper
   * @return fixedRate Fixed rate wrapper deployed
   */

  function _deployFixedRate(
    ISynthereumFactoryVersioning factoryVersioning,
    uint8 fixedRateVersion,
    bytes memory fixedRateParamsData
  ) internal returns (ISynthereumDeployment fixedRate) {
    address fixedRateFactory =
      factoryVersioning.getFactoryVersion(
        FactoryInterfaces.FixedRateFactory,
        fixedRateVersion
      );
    bytes memory fixedRateDeploymentResult =
      fixedRateFactory.functionCall(
        abi.encodePacked(
          getDeploymentSignature(fixedRateFactory),
          fixedRateParamsData
        ),
        'Wrong fixed rate deployment'
      );
    fixedRate = ISynthereumDeployment(
      abi.decode(fixedRateDeploymentResult, (address))
    );
  }

  /**
   * @notice Sets roles of the synthetic token contract to a pool or a fixed rate wrapper
   * @param pool Pool or fixed rate wrapper contract
   */
  function setSyntheticTokenRoles(ISynthereumDeployment pool) internal {
    address _pool = address(pool);
    IAccessControlEnumerable tokenCurrency =
      IAccessControlEnumerable(address(pool.syntheticToken()));
    if (
      !tokenCurrency.hasRole(MINTER_ROLE, _pool) ||
      !tokenCurrency.hasRole(BURNER_ROLE, _pool)
    ) {
      addSyntheticTokenRoles(address(tokenCurrency), _pool);
    }
  }

  /**
   * @notice Grants minter and burner role of syntehtic token to derivative
   * @param tokenCurrency Address of the token contract
   * @param contractAddr Address of the pool or self-minting derivative
   */
  function addSyntheticTokenRoles(address tokenCurrency, address contractAddr)
    internal
  {
    ISynthereumManager manager = getManager();
    address[] memory contracts = new address[](2);
    bytes32[] memory roles = new bytes32[](2);
    address[] memory accounts = new address[](2);
    contracts[0] = tokenCurrency;
    contracts[1] = tokenCurrency;
    roles[0] = MINTER_ROLE;
    roles[1] = BURNER_ROLE;
    accounts[0] = contractAddr;
    accounts[1] = contractAddr;
    manager.grantSynthereumRole(contracts, roles, accounts);
  }

  //----------------------------------------
  // Internal view functions
  //----------------------------------------

  /**
   * @notice Get factory versioning contract from the finder
   * @return factoryVersioning Factory versioning contract
   */
  function getFactoryVersioning()
    internal
    view
    returns (ISynthereumFactoryVersioning factoryVersioning)
  {
    factoryVersioning = ISynthereumFactoryVersioning(
      synthereumFinder.getImplementationAddress(
        SynthereumInterfaces.FactoryVersioning
      )
    );
  }

  /**
   * @notice Get pool registry contract from the finder
   * @return poolRegistry Registry of pools
   */
  function getPoolRegistry()
    internal
    view
    returns (ISynthereumRegistry poolRegistry)
  {
    poolRegistry = ISynthereumRegistry(
      synthereumFinder.getImplementationAddress(
        SynthereumInterfaces.PoolRegistry
      )
    );
  }

  /**
   * @notice Get self minting registry contract from the finder
   * @return selfMintingRegistry Registry of self-minting derivatives
   */
  function getSelfMintingRegistry()
    internal
    view
    returns (ISynthereumRegistry selfMintingRegistry)
  {
    selfMintingRegistry = ISynthereumRegistry(
      synthereumFinder.getImplementationAddress(
        SynthereumInterfaces.SelfMintingRegistry
      )
    );
  }

  /**
   * @notice Get fixed rate registry contract from the finder
   * @return fixedRateRegistry Registry of fixed rate contract
   */
  function getFixedRateRegistry()
    internal
    view
    returns (ISynthereumRegistry fixedRateRegistry)
  {
    fixedRateRegistry = ISynthereumRegistry(
      synthereumFinder.getImplementationAddress(
        SynthereumInterfaces.FixedRateRegistry
      )
    );
  }

  /**
   * @notice Get manager contract from the finder
   * @return manager Synthereum manager
   */
  function getManager() internal view returns (ISynthereumManager manager) {
    manager = ISynthereumManager(
      synthereumFinder.getImplementationAddress(SynthereumInterfaces.Manager)
    );
  }

  /**
   * @notice Get signature of function to deploy a contract
   * @param factory Factory contract
   * @return signature Signature of deployment function of the factory
   */
  function getDeploymentSignature(address factory)
    internal
    view
    returns (bytes4 signature)
  {
    signature = IDeploymentSignature(factory).deploymentSignature();
  }

  /**
   * @notice Check correct finder and version of the deployed pool or self-minting derivative
   * @param poolOrDerivative Contract pool or self-minting derivative to check
   * @param version Pool or self-minting derivative version to check
   */
  function checkDeployment(
    ISynthereumDeployment poolOrDerivative,
    uint8 version
  ) internal view {
    require(
      poolOrDerivative.synthereumFinder() == synthereumFinder,
      'Wrong finder in deployment'
    );
    require(
      poolOrDerivative.version() == version,
      'Wrong version in deployment'
    );
  }
}
