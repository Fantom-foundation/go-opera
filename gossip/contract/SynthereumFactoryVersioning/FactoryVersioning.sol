// SPDX-License-Identifier: AGPL-3.0-only
pragma solidity ^0.8.4;

import {
  ISynthereumFactoryVersioning
} from './interfaces/IFactoryVersioning.sol';
import {
  EnumerableMap
} from './@openzeppelin/contracts/utils/structs/EnumerableMap.sol';
import {
  AccessControlEnumerable
} from './@openzeppelin/contracts/access/AccessControlEnumerable.sol';

/**
 * @title Provides addresses of different versions of pools factory and derivative factory
 */
contract SynthereumFactoryVersioning is
  ISynthereumFactoryVersioning,
  AccessControlEnumerable
{
  using EnumerableMap for EnumerableMap.UintToAddressMap;

  bytes32 public constant MAINTAINER_ROLE = keccak256('Maintainer');

  //Describe role structure
  struct Roles {
    address admin;
    address maintainer;
  }

  //----------------------------------------
  // Storage
  //----------------------------------------

  mapping(bytes32 => EnumerableMap.UintToAddressMap) private factories;

  //----------------------------------------
  // Events
  //----------------------------------------

  event AddFactory(
    bytes32 indexed factoryType,
    uint8 indexed version,
    address indexed factory
  );

  event SetFactory(
    bytes32 indexed factoryType,
    uint8 indexed version,
    address indexed factory
  );

  event RemoveFactory(
    bytes32 indexed factoryType,
    uint8 indexed version,
    address indexed factory
  );

  //----------------------------------------
  // Constructor
  //----------------------------------------
  constructor(Roles memory roles) {
    _setRoleAdmin(DEFAULT_ADMIN_ROLE, DEFAULT_ADMIN_ROLE);
    _setRoleAdmin(MAINTAINER_ROLE, DEFAULT_ADMIN_ROLE);
    _setupRole(DEFAULT_ADMIN_ROLE, roles.admin);
    _setupRole(MAINTAINER_ROLE, roles.maintainer);
  }

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
  // External functions
  //----------------------------------------

  /** @notice Sets a Factory
   * @param factoryType Type of factory
   * @param version Version of the factory to be set
   * @param factory The pool factory address to be set
   */
  function setFactory(
    bytes32 factoryType,
    uint8 version,
    address factory
  ) external override onlyMaintainer {
    require(factory != address(0), 'Factory cannot be address 0');
    bool isNewVersion = factories[factoryType].set(version, factory);
    if (isNewVersion) {
      emit AddFactory(factoryType, version, factory);
    } else {
      emit SetFactory(factoryType, version, factory);
    }
  }

  /** @notice Removes a factory
   * @param factoryType The type of factory to be removed
   * @param version Version of the factory to be removed
   */
  function removeFactory(bytes32 factoryType, uint8 version)
    external
    override
    onlyMaintainer
  {
    EnumerableMap.UintToAddressMap storage selectedFactories =
      factories[factoryType];
    address factoryToRemove = selectedFactories.get(version);
    selectedFactories.remove(version);
    emit RemoveFactory(factoryType, version, factoryToRemove);
  }

  //----------------------------------------
  // External view functions
  //----------------------------------------

  /** @notice Gets a factory contract address
   * @param factoryType The type of factory to be checked
   * @param version Version of the factory to be checked
   * @return factory Address of the factory contract
   */
  function getFactoryVersion(bytes32 factoryType, uint8 version)
    external
    view
    override
    returns (address factory)
  {
    factory = factories[factoryType].get(version);
  }

  /** @notice Gets the number of factory versions for a specific type
   * @param factoryType The type of factory to be checked
   * @return numberOfVersions Total number of versions for a specific factory
   */
  function numberOfFactoryVersions(bytes32 factoryType)
    external
    view
    override
    returns (uint8 numberOfVersions)
  {
    numberOfVersions = uint8(factories[factoryType].length());
  }
}

